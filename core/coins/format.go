package coins

import (
	"fmt"
	"sort"
	"strings"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
)

// DefaultSeparator is the default separator used between formatted coins
const DefaultSeparator = ", "

// ErrMetadataMismatch is returned when the number of coins doesn't match the number of metadata entries
var ErrMetadataMismatch = fmt.Errorf("number of metadata entries must match number of coins")

// formatCoin formats a single coin with its metadata into a human-readable string.
// It returns the formatted string and any error encountered during formatting.
func formatCoin(coin *basev1beta1.Coin, metadata *bankv1beta1.Metadata) (string, error) {
	if coin == nil {
		return "", fmt.Errorf("nil coin")
	}

	// Handle cases without metadata or display denom
	if shouldUseOriginalDenom(coin.Denom, metadata) {
		return formatOriginalCoin(coin)
	}

	return formatWithMetadata(coin, metadata)
}

// FormatCoins formats multiple coins with their metadata into a sorted, human-readable string.
// The metadata slice must have the same length as the coins slice, with matching indices.
func FormatCoins(coins []*basev1beta1.Coin, metadata []*bankv1beta1.Metadata) (string, error) {
	if len(coins) != len(metadata) {
		return "", fmt.Errorf("%w: expected %d, got %d", 
			ErrMetadataMismatch, len(coins), len(metadata))
	}

	if len(coins) == 0 {
		return "", nil
	}

	formatted, err := formatAllCoins(coins, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to format coins: %w", err)
	}

	sortFormattedCoins(formatted)
	return strings.Join(formatted, DefaultSeparator), nil
}

// Helper functions

func shouldUseOriginalDenom(coinDenom string, metadata *bankv1beta1.Metadata) bool {
	return metadata == nil || metadata.Display == "" || coinDenom == metadata.Display
}

func formatOriginalCoin(coin *basev1beta1.Coin) (string, error) {
	vr, err := math.FormatDec(coin.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to format amount: %w", err)
	}
	return fmt.Sprintf("%s %s", vr, coin.Denom), nil
}

func formatWithMetadata(coin *basev1beta1.Coin, metadata *bankv1beta1.Metadata) (string, error) {
	coinExp, dispExp, err := findExponents(coin.Denom, metadata.Display, metadata.DenomUnits)
	if err != nil {
		return formatOriginalCoin(coin)
	}

	dispAmount, err := calculateDisplayAmount(coin.Amount, coinExp, dispExp)
	if err != nil {
		return "", fmt.Errorf("failed to calculate display amount: %w", err)
	}

	vr, err := math.FormatDec(dispAmount.String())
	if err != nil {
		return "", fmt.Errorf("failed to format display amount: %w", err)
	}

	return fmt.Sprintf("%s %s", vr, metadata.Display), nil
}

func findExponents(coinDenom, dispDenom string, units []*bankv1beta1.DenomUnit) (coinExp, dispExp uint32, err error) {
	var foundCoin, foundDisp bool

	for _, unit := range units {
		switch unit.Denom {
		case coinDenom:
			coinExp = unit.Exponent
			foundCoin = true
		case dispDenom:
			dispExp = unit.Exponent
			foundDisp = true
		}
	}

	if !foundCoin || !foundDisp {
		return 0, 0, fmt.Errorf("exponents not found")
	}

	return coinExp, dispExp, nil
}

func calculateDisplayAmount(amount string, coinExp, dispExp uint32) (math.LegacyDec, error) {
	dispAmount, err := math.LegacyNewDecFromStr(amount)
	if err != nil {
		return math.LegacyDec{}, fmt.Errorf("invalid amount: %w", err)
	}

	power := math.LegacyNewDec(10)
	if coinExp > dispExp {
		return dispAmount.Mul(power.Power(uint64(coinExp - dispExp))), nil
	}
	return dispAmount.Quo(power.Power(uint64(dispExp - coinExp))), nil
}

func formatAllCoins(coins []*basev1beta1.Coin, metadata []*bankv1beta1.Metadata) ([]string, error) {
	formatted := make([]string, len(coins))
	for i, coin := range coins {
		var err error
		formatted[i], err = formatCoin(coin, metadata[i])
		if err != nil {
			return nil, fmt.Errorf("failed to format coin at index %d: %w", i, err)
		}
	}
	return formatted, nil
}

func sortFormattedCoins(formatted []string) {
	sort.SliceStable(formatted, func(i, j int) bool {
		denomI := strings.Split(formatted[i], " ")[1]
		denomJ := strings.Split(formatted[j], " ")[1]
		return denomI < denomJ
	})
}
