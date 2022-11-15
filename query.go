package main

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

// Used by: Zacks rank, zacks industry rank
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
