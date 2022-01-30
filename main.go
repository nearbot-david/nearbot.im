package main

import (
	goSQL "database/sql"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"github.com/mazanax/moneybot/endpoints"
	"github.com/mazanax/moneybot/handlers"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/services"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func isDebug() bool {
	return os.Getenv("DEBUG") == "1"
}

func getDb() *goqu.Database {
	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		log.Panic("DB_URI is required")
	}
	uri, err := pq.ParseURL(dbURI)
	if err != nil {
		panic(err)
	}
	pdb, err := goSQL.Open("postgres", uri)
	if err != nil {
		panic(err)
	}
	return goqu.New("postgres", pdb)
}

func main() {
	bot, err := tg.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	db := getDb()
	balanceRepository := repository.NewBalanceRepository(db)
	depositRepository := repository.NewDepositRepository(db)
	transferRepository := repository.NewTransferRepository(db)
	stateRepository := repository.NewStateRepository(db)
	withdrawalRepository := repository.NewWithdrawalRepository(db)
	historyRepository := repository.NewHistoryRepository(db)

	historyManager := services.NewHistoryManager(historyRepository)
	withdrawalManager := services.NewWithdrawalManager(withdrawalRepository)
	stateManager := services.NewStateManager(stateRepository)
	balanceManager := services.NewBalanceManager(balanceRepository, transferRepository)
	paymentMethod := services.NewGatewayPaymentMethod(
		os.Getenv("PAYMENT_ENDPOINT"),
		os.Getenv("PAYMENT_SUCCESS_ENDPOINT"),
		os.Getenv("GATEWAY_CLIENT_ID"),
		os.Getenv("GATEWAY_SECRET_KEY"),
		os.Getenv("GATEWAY_HOST"),
		depositRepository,
	)

	bot.Debug = isDebug()
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", endpoints.IndexEndpoint())
		mux.HandleFunc("/faq", endpoints.FaqEndpoint())
		mux.HandleFunc("/payment/", endpoints.PaymentEndpoint(paymentMethod, depositRepository, balanceManager, historyManager, bot, isDebug()))
		mux.HandleFunc("/payment-successful", endpoints.PaymentSuccessfulEndpoint())
		mux.HandleFunc("/payment-failed", endpoints.PaymentFailedEndpoint())
		if isDebug() {
			mux.HandleFunc("/emulator/", endpoints.EmulatorEndpoint(depositRepository, os.Getenv("GATEWAY_SECRET_KEY")))
		}

		err = http.ListenAndServe(fmt.Sprintf(":%d", 8444), mux)
		if err != nil {
			fmt.Printf("Cannot start http server: %s\n", err)
			os.Exit(2)
		}
	}()

	for update := range updates {
		switch true {
		case update.Message != nil && update.Message.Command() != "":
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			switch update.Message.Command() {
			case "start":
				handlers.HandleStart(balanceManager)(bot, &update)
			default:
				// unknown command
			}
		case update.Message != nil:
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			switch stateManager.GetState(update.Message.From.ID) {
			case models.UserStateWithdrawAmount:
				handlers.HandleWithdrawAmount(balanceManager, stateManager, withdrawalManager)(bot, &update)
			case models.UserStateWithdrawAddress:
				handlers.HandleWithdrawAddress(balanceManager, stateManager, withdrawalManager)(bot, &update)
			default:
				handlers.HandleUnsupportedMessage()(bot, &update)
			}
		case update.CallbackQuery != nil:
			log.Printf("[%s] Callback query %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
			switch true {
			case update.CallbackData() == "support":
				handlers.HandleSupport()(bot, &update)
			case update.CallbackData() == "cancel":
				handlers.HandleCancel(balanceManager, stateManager)(bot, &update)
			case update.CallbackData() == "show_balance":
				handlers.HandleBalance(balanceManager)(bot, &update)
			case update.CallbackData() == "history":
				handlers.HandleHistory()(bot, &update)
			case update.CallbackData() == "deposit":
				handlers.HandleDeposit(balanceManager)(bot, &update)
			case update.CallbackData() == "withdraw":
				handlers.HandleWithdraw(balanceManager, stateManager, withdrawalManager)(bot, &update)
			case update.CallbackData() == "withdraw_confirm":
				handlers.HandleWithdrawConfirm(balanceManager, stateManager, withdrawalManager, historyManager)(bot, &update)
			case update.CallbackData() == "withdraw_cancel":
				handlers.HandleWithdrawCancel(withdrawalManager)(bot, &update)
			case update.CallbackData() == "withdraw_cancel_confirm":
				handlers.HandleWithdrawCancelConfirm(balanceManager, withdrawalManager, historyManager)(bot, &update)
			case strings.HasPrefix(update.CallbackData(), "deposit_"):
				amountString := strings.TrimPrefix(update.CallbackData(), "deposit_")
				amount, _ := strconv.Atoi(amountString)

				handlers.HandleDepositAmount(paymentMethod, depositRepository, uint64(amount))(bot, &update)
			case strings.HasPrefix(update.CallbackData(), "transfer_approve_"):
				transferSlug := strings.TrimPrefix(update.CallbackData(), "transfer_approve_")

				handlers.HandleTransferApprove(balanceManager, transferRepository, transferSlug, historyManager)(bot, &update)
			case strings.HasPrefix(update.CallbackData(), "transfer_reject_"):
				transferSlug := strings.TrimPrefix(update.CallbackData(), "transfer_reject_")

				handlers.HandleTransferReject(balanceManager, transferRepository, transferSlug, historyManager)(bot, &update)
			}
		case update.InlineQuery != nil:
			handlers.HandleTransfer(balanceManager)(bot, &update)
		case update.ChosenInlineResult != nil:
			handlers.HandleTransferSent(balanceManager, historyManager)(bot, &update)
		default:
			continue
		}
	}
}
