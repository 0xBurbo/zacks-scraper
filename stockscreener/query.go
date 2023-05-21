package stockscreener

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
)

// Parses config to write screener query
func WriteQuery(w *multipart.Writer, config []map[string]interface{}) error {

	writeStockScreenerBaseQuery(w)

	for _, item := range config {
		id, ok := item["id"].(string)
		if !ok {
			return errors.New("each item in query list must have 'id' field")
		}

		value, ok := item["value"].(string)
		if !ok {
			return errors.New("each item in query list must have 'value' field")
		}

		operator, ok := item["operator"].(string)
		if !ok {
			return errors.New("each item in query list must have 'operator' field")
		}

		switch id {
		case "zacks_rank":
			numValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("unable to parse int from zacks_rank query value %v: %w", value, err)
			}

			err = writeZacksRankQuery(w, numValue, operator)

			if err != nil {
				return err
			}
		case "zacks_industry_rank":
			numValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("unable to parse int from zacks_rank query value %v: %w", value, err)
			}

			err = writeZacksIndustryRankQuery(w, numValue, operator)

			if err != nil {
				return err
			}
		case "value_score":
			err := writeValueScoreQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "growth_score":
			err := writeGrowthScoreQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "momentum_score":
			err := writeMomentumScoreQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "vgm_score":
			err := writeVGMScoreQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "earnings_esp":
			err := writeEarningsESPQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "52_week_high":
			err := write52WeekHighQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "market_cap":
			err := writeMarketCapQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "last_eps_surprise":
			err := writeLastEPSSurpriseQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "p_n_e":
			err := writePoverEQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "num_brokers":
			err := writeNumBrokersQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "optionable":
			err := writeOptionableQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "percent_change_f1":
			err := writeChangeF1Query(w, value, operator)
			if err != nil {
				return err
			}
		case "div_yield":
			err := writeDivYieldQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "avg_volume":
			err := writeAvgVolumeQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "last_eps_report_date":
			err := writeLastEpsReportDateQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "next_eps_report_date":
			err := writeNextEpsReportDateQuery(w, value, operator)
			if err != nil {
				return err
			}
		case "q0_consensus_est":
			err := writeQ0ConsensusEst(w, value, operator)
			if err != nil {
				return err
			}
		case "last_reported_quarter":
			err := writeLastReportedQtrQuery(w, value, operator)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// Operators are represented as integers in form data
// Only used for Value/VGM/Growth/Momentum
var operatorCodesForScores = map[string]int{
	">=": 12,
	"<=": 13,
	"=":  19,
	"<>": 20,
}

// Used by: Zacks rank, zacks industry rank, last EPS report date
var zackRankOperatorMap = map[string]int{
	">=": 6,
	"<=": 7,
	"=":  8,
	"<>": 17,
}

func isValidGrade(grade string) bool {
	grades := []string{"A", "B", "C", "D", "F"}
	for _, v := range grades {
		if v == grade {
			return true
		}
	}
	return false
}

func writeStockScreenerBaseQuery(writer *multipart.Writer) {
	// Required query params
	writer.WriteField("is_only_matches", "0")
	writer.WriteField("is_premium_exists", "0")
	writer.WriteField("is_edit_view", "0")
	writer.WriteField("saved_screen_name", "")
	writer.WriteField("tab_id", "1")
	writer.WriteField("start_page", "1")
	writer.WriteField("no_of_rec", "15")
	writer.WriteField("sort_col", "2")
	writer.WriteField("sort_type", "ASC")

	// "My Criteria"

}

func writeZacksRankQuery(writer *multipart.Writer, value int, operator string) error {
	if value < 1 || value > 5 {
		return errors.New("zacks rank value must be between 1 and 5")
	}
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for zacks rank: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", fmt.Sprintf("%v", value))
	writer.WriteField("p_items[]", "15005") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Zacks Rank")
	writer.WriteField("p_item_key[]", "0")
	return nil
}

func writeZacksIndustryRankQuery(writer *multipart.Writer, value int, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for zacks industry rank: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", fmt.Sprintf("%v", value))
	writer.WriteField("p_items[]", "15025") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Zacks Industry Rank")
	writer.WriteField("p_item_key[]", "1")
	return nil
}

func writeValueScoreQuery(writer *multipart.Writer, grade, operator string) error {
	operatorCode, ok := operatorCodesForScores[operator]
	if !ok {
		return errors.New("Unknown operator for value score: " + operator)
	}

	if !isValidGrade(grade) {
		return errors.New("Invalid grade param for value score: " + grade)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", grade)
	writer.WriteField("p_items[]", "15030") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Value Score")
	writer.WriteField("p_item_key[]", "2")
	return nil
}

func writeGrowthScoreQuery(writer *multipart.Writer, grade, operator string) error {
	operatorCode, ok := operatorCodesForScores[operator]
	if !ok {
		return errors.New("Unknown operator for growth score: " + operator)
	}

	if !isValidGrade(grade) {
		return errors.New("Invalid grade param for growth score: " + grade)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", grade)
	writer.WriteField("p_items[]", "15035") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Growth Score")
	writer.WriteField("p_item_key[]", "3")
	return nil
}

func writeMomentumScoreQuery(writer *multipart.Writer, grade, operator string) error {
	operatorCode, ok := operatorCodesForScores[operator]
	if !ok {
		return errors.New("Unknown operator for momentum score: " + operator)
	}

	if !isValidGrade(grade) {
		return errors.New("Invalid grade param for momentum score: " + grade)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", grade)
	writer.WriteField("p_items[]", "15040") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Momentum Score")
	writer.WriteField("p_item_key[]", "4")
	return nil
}

func writeVGMScoreQuery(writer *multipart.Writer, grade, operator string) error {
	operatorCode, ok := operatorCodesForScores[operator]
	if !ok {
		return errors.New("Unknown operator for VGM score: " + operator)
	}

	if !isValidGrade(grade) {
		return errors.New("Invalid grade param for VGM score: " + grade)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", grade)
	writer.WriteField("p_items[]", "15045") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "VGM Score")
	writer.WriteField("p_item_key[]", "5")
	return nil
}

func writeEarningsESPQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for earnings esp query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "17060") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Earnings ESP")
	writer.WriteField("p_item_key[]", "6")
	return nil
}

func write52WeekHighQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for 52 week high query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "14010") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "52 Week High")
	writer.WriteField("p_item_key[]", "7")
	return nil
}

func writeMarketCapQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for market cap query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "12010") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Market Cap (mil)")
	writer.WriteField("p_item_key[]", "8")
	return nil
}

func writeLastEPSSurpriseQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for EPS surprise query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "17005") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Last EPS Surprise (%)")
	writer.WriteField("p_item_key[]", "9")
	return nil
}

func writePoverEQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for P/E query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "22010") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "P/E (F1)")
	writer.WriteField("p_item_key[]", "10")
	return nil
}

func writeNumBrokersQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for # of brokers query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "16010") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "# of Brokers in Rating")
	writer.WriteField("p_item_key[]", "11")
	return nil
}

func writeOptionableQuery(writer *multipart.Writer, value, operator string) error {
	var op int = -1
	if operator == "EQUAL" {
		op = 9
	} else if operator == "NOT EQUAL" {
		op = 18
	} else {
		return errors.New("operator for optionable must be either EQUAL or NOT EQUAL")
	}

	if value != "NO" && value != "YES" {
		return errors.New("value for optionable must be either NO or YES")
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", op))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "11015") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Optionable")
	writer.WriteField("p_item_key[]", "12")
	return nil
}

func writeChangeF1Query(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for percent change F1 query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "18020") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "% Change F1 Est. (4 weeks)")
	writer.WriteField("p_item_key[]", "13")
	return nil
}

func writeDivYieldQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for div yield query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "25005") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Div. Yield %")
	writer.WriteField("p_item_key[]", "14")
	return nil
}

func writeAvgVolumeQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for avg volume query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "12015") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Avg Volume")
	writer.WriteField("p_item_key[]", "15")
	return nil
}

func writeLastEpsReportDateQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for last EPS report date query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "17050") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Last EPS Report Date (yyyymmdd)")
	writer.WriteField("p_item_key[]", "72")
	return nil
}

func writeNextEpsReportDateQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for next EPS report date query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "17055") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Next EPS Report Date (yyyymmdd)")
	writer.WriteField("p_item_key[]", "73")
	return nil
}

func writeQ0ConsensusEst(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for Q0 consensus estimate query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "19005") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Q0 Consensus Est. (last completed fiscal Qtr)")
	writer.WriteField("p_item_key[]", "80")
	return nil
}

func writeLastReportedQtrQuery(writer *multipart.Writer, value, operator string) error {
	operatorCode, ok := zackRankOperatorMap[operator]
	if !ok {
		return errors.New("Unknown operator for last reported quarter query: " + operator)
	}

	writer.WriteField("operator[]", fmt.Sprintf("%v", operatorCode))
	writer.WriteField("value[]", value)
	writer.WriteField("p_items[]", "17030") // TODO: Retrieve this value from zacks
	writer.WriteField("p_item_name[]", "Last Reported Qtr (yyyymm)")
	writer.WriteField("p_item_key[]", "68")
	return nil
}
