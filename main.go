package main

import (
	goSQL "database/sql"
	"fmt"
	"github.com/Pay-With-NEAR/nearbot.im/endpoints"
	"github.com/Pay-With-NEAR/nearbot.im/handlers"
	"github.com/Pay-With-NEAR/nearbot.im/models"
	"github.com/Pay-With-NEAR/nearbot.im/repository"
	"github.com/Pay-With-NEAR/nearbot.im/services"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func isDebug() bool {
	return os.Getenv("DEBUG") == "1"
}

func shouldRunCron() bool {
	return os.Getenv("NO_CRON") != "1"
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
	transactionRepository := repository.NewTransactionRepository(db)

	addressManager := services.NewAddressManager(balanceRepository)
	historyManager := services.NewHistoryManager(historyRepository)
	withdrawalManager := services.NewWithdrawalManager(withdrawalRepository)
	stateManager := services.NewStateManager(stateRepository)
	balanceManager := services.NewBalanceManager(balanceRepository, transferRepository, addressManager)
	paymentMethod := services.NewPaywithnearMethod(
		os.Getenv("PAYMENT_ENDPOINT"),
		os.Getenv("PWN_CLIENT_ID"),
		os.Getenv("PWN_CLIENT_SECRET"),
		os.Getenv("PWN_HOST"),
		os.Getenv("PAYMENT_SUCCESS_ENDPOINT"),
		depositRepository,
	)
	transactionChecker := services.NewTransactionChecker(db, balanceRepository, transactionRepository)

	bot.Debug = isDebug()
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	if shouldRunCron() {
		go func(bot *tg.BotAPI) {
			transactionChecker.Run(bot)
		}(bot)
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", endpoints.IndexEndpoint())
		mux.HandleFunc("/faq", endpoints.FaqEndpoint())
		mux.HandleFunc("/payment/", endpoints.PaymentEndpoint(paymentMethod, balanceManager, historyManager, addressManager, bot))
		mux.HandleFunc("/payment-successful", endpoints.PaymentSuccessfulEndpoint())
		mux.HandleFunc("/payment-success", endpoints.PaymentSuccessfulEndpoint())
		mux.HandleFunc("/payment-failed", endpoints.PaymentFailedEndpoint())
		mux.HandleFunc("/address/", endpoints.QrCodeEndpoint())
		if isDebug() {
			mux.HandleFunc("/emulator/", endpoints.EmulatorEndpoint(depositRepository, os.Getenv("GATEWAY_SECRET_KEY")))
		}

		serverPort := os.Getenv("SERVER_PORT")
		if serverPort == "" {
			serverPort = "8444"
		}

		serverPortInt, _ := strconv.Atoi(serverPort)

		err = http.ListenAndServe(fmt.Sprintf(":%d", serverPortInt), mux)
		if err != nil {
			fmt.Printf("Cannot start http server: %s\n", err)
			os.Exit(2)
		}
	}()

	for update := range updates {
		go func(update tg.Update) {
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
					if update.Message != nil && len(update.Message.NewChatMembers) > 0 {
						return
					}
					if update.Message.ViaBot != nil &&
						update.Message.ViaBot.UserName == bot.Self.UserName &&
						update.Message.ReplyToMessage != nil {
						handlers.HandleTransferReceiver(transferRepository)(bot, &update)
						return
					}
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
				case update.CallbackData() == "show_address":
					handlers.HandleAddress(balanceManager)(bot, &update)
				case update.CallbackData() == "history":
					handlers.HandleHistory()(bot, &update)
				case update.CallbackData() == "deposit":
					handlers.HandleDeposit(balanceManager)(bot, &update)
				case update.CallbackData() == "withdraw":
					handlers.HandleWithdraw(balanceManager, stateManager, withdrawalManager)(bot, &update)
				case update.CallbackData() == "withdraw_confirm":
					handlers.HandleWithdrawConfirm(balanceManager, stateManager, withdrawalManager, historyManager, addressManager)(bot, &update)
				case strings.HasPrefix(update.CallbackData(), "deposit_"):
					amountString := strings.TrimPrefix(update.CallbackData(), "deposit_")
					amount, _ := strconv.Atoi(amountString)

					handlers.HandleDepositAmount(paymentMethod, depositRepository, uint64(amount))(bot, &update)
				case strings.HasPrefix(update.CallbackData(), "transfer_approve_"):
					transferSlug := strings.TrimPrefix(update.CallbackData(), "transfer_approve_")

					handlers.HandleTransferApprove(balanceManager, transferRepository, transferSlug, historyManager, addressManager)(bot, &update)
				case strings.HasPrefix(update.CallbackData(), "transfer_reject_"):
					transferSlug := strings.TrimPrefix(update.CallbackData(), "transfer_reject_")

					handlers.HandleTransferReject(balanceManager, transferRepository, transferSlug, historyManager)(bot, &update)
				}
			case update.InlineQuery != nil:
				handlers.HandleTransfer(balanceManager)(bot, &update)
			case update.ChosenInlineResult != nil:
				handlers.HandleTransferSent(balanceManager, historyManager, transferRepository)(bot, &update)
			default:
				return
			}
		}(update)
	}
}
