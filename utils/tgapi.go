package utils

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type CachelessInlineConfig struct {
	InlineQueryID     string        `json:"inline_query_id"`
	Results           []interface{} `json:"results"`
	CacheTime         int           `json:"cache_time"`
	IsPersonal        bool          `json:"is_personal"`
	NextOffset        string        `json:"next_offset"`
	SwitchPMText      string        `json:"switch_pm_text"`
	SwitchPMParameter string        `json:"switch_pm_parameter"`
}

func (config CachelessInlineConfig) Params() (tg.Params, error) {
	params := make(tg.Params)

	params["inline_query_id"] = config.InlineQueryID
	params["cache_time"] = "0"
	params.AddBool("is_personal", config.IsPersonal)
	params.AddNonEmpty("next_offset", config.NextOffset)
	params.AddNonEmpty("switch_pm_text", config.SwitchPMText)
	params.AddNonEmpty("switch_pm_parameter", config.SwitchPMParameter)
	err := params.AddInterface("results", config.Results)

	return params, err
}

func (config CachelessInlineConfig) Method() string {
	return "answerInlineQuery"
}
