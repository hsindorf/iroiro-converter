package amountconverter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hsindorf/iroiro-converter/counting"
	"github.com/hsindorf/iroiro-converter/currencyconverter"
	"github.com/hsindorf/iroiro-converter/distanceconverter"
	"github.com/hsindorf/iroiro-converter/stringutils"
	"golang.org/x/exp/slices"
)

// Convert takes an amount - either a numerical amount or a currency amount, and:
//   - if it's a currency, converts the currency, and formats based on the target provided. e.g. ($100000, "jp") => "1万円"
//   - if it's only a number, converts the number to the alternative format. Ignores the target and flag.
//
// at the moment, only supports formats "12345", "$12345", or "12345円"
func Convert(amount string, rate float64, useJPUnits bool) (string, error) {
	currency, parsedAmount := ParseAmount(amount)

	num, err := counting.Parse(parsedAmount)
	if err != nil {
		return "", err
	}

	if currency == "" {
		if counting.IsJapaneseNumber((parsedAmount)) {
			return fmt.Sprintf("%v", stringutils.Commafy(num)), nil
		} else {
			return counting.ConvertToLargestUnit(num), nil
		}
	}

	if currency == "$" {
		amountInYen := currencyconverter.DollarsToYen(num, rate)
		if useJPUnits {
			return fmt.Sprintf("%v円", counting.ConvertToLargestUnit(amountInYen)), nil
		}
		return fmt.Sprintf("%v円", stringutils.Commafy(amountInYen)), nil
	}

	if currency == "円" {
		amountInDollars := currencyconverter.YenToDollars(num, rate)
		if useJPUnits {
			return fmt.Sprintf("$%v", counting.ConvertToLargestUnit(amountInDollars)), nil
		}
		return fmt.Sprintf("$%v", stringutils.Commafy(amountInDollars)), nil
	}

	if slices.Contains([]string{"cm", "in", "m", "ft", "km", "mi"}, currency) {
		return distanceconverter.ConvertDistance(currency, num, useJPUnits), nil
	}

	return "", errors.New("something bad happened")
}

// ParseAmount splits an input into currency (if present) and amount
//   - e.g. "$100" => ("$", "100")
//   - e.g. "100 yen" => ("円", "100")
func ParseAmount(amount string) (string, string) {
	splitAmount := strings.Split(amount, " ")

	if len(splitAmount) > 1 && splitAmount[len(splitAmount)-1] == "dollars" {
		return "$", strings.Join(splitAmount[:len(splitAmount)-1], " ")
	}

	if len(splitAmount) > 1 && splitAmount[len(splitAmount)-1] == "yen" {
		return "円", strings.Join(splitAmount[:len(splitAmount)-1], " ")
	}

	runeAmount := []rune(amount)

	if len(runeAmount) > 2 {
		twoCharSuffix := string(runeAmount[len(runeAmount)-2]) + string(runeAmount[len(runeAmount)-1])
		if twoCharSuffix == "ドル" {
			return "$", string(runeAmount[:len(runeAmount)-2])
		}
		if twoCharSuffix == "cm" ||
			twoCharSuffix == "in" ||
			twoCharSuffix == "ft" ||
			twoCharSuffix == "km" ||
			twoCharSuffix == "mi" {
			return twoCharSuffix, string(runeAmount[:len(runeAmount)-2])
		}
	}

	if string(runeAmount[0]) == "$" ||
		string(runeAmount[0]) == "＄" {
		return "$", string(runeAmount[1:])
	}

	if string(runeAmount[len(runeAmount)-1]) == "円" {
		return "円", string(runeAmount[:len(runeAmount)-1])
	}

	if string(runeAmount[len(runeAmount)-1]) == "m" {
		return "m", string(runeAmount[:len(runeAmount)-1])
	}

	return "", amount
}