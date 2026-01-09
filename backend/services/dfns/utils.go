package dfns

import (
	"math/big"
	"regexp"
	"strings"
)

var evmAddressRegex = regexp.MustCompile("^0x[a-fA-F0-9]{40}$")
var tronAddressRegex = regexp.MustCompile("^T[a-zA-Z0-9]{33}$")

// ValidChains contains the valid chain names
var ValidChains = map[string]bool{
	"ethereum":         true,
	"ethereum-sepolia": true,
	"tron":             true,
	"tron-nile":        true,
}

// ValidTokens contains the valid token symbols
var ValidTokens = map[string]bool{
	"USDC": true,
	"USDT": true,
}

// IsValidEVMAddress validates an EVM address format
func IsValidEVMAddress(address string) bool {
	return evmAddressRegex.MatchString(address)
}

// IsValidTronAddress validates a TRON address format
func IsValidTronAddress(address string) bool {
	return tronAddressRegex.MatchString(address)
}

// IsValidAddress validates an address for a given chain
func IsValidAddress(address string, chainName string) bool {
	if strings.HasPrefix(chainName, "tron") {
		return IsValidTronAddress(address)
	}
	return IsValidEVMAddress(address)
}

// IsValidChainName validates a chain name
func IsValidChainName(chain string) bool {
	return ValidChains[chain]
}

// IsTronChain checks if the chain is TRON-based
func IsTronChain(chainName string) bool {
	return strings.HasPrefix(chainName, "tron")
}

// IsTestnet checks if the chain is a testnet
func IsTestnet(chainName string) bool {
	return strings.Contains(chainName, "sepolia") || strings.Contains(chainName, "nile")
}

// IsValidTokenSymbol validates a token symbol
func IsValidTokenSymbol(symbol string) bool {
	return ValidTokens[symbol]
}

// ConvertToCredits converts a raw token amount to credits
// For USDC/USDT (6 decimals): 1,000,000 raw units = 1 credit
func ConvertToCredits(rawAmount string, decimals int) int64 {
	amount := new(big.Int)
	amount.SetString(rawAmount, 10)

	// Create divisor: 10^decimals
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// Divide to get credits (integer division, truncates fractional credits)
	credits := new(big.Int).Div(amount, divisor)

	return credits.Int64()
}

// CreditsToTokenAmount converts credits to raw token amount
// For USDC/USDT (6 decimals): 1 credit = 1,000,000 raw units
func CreditsToTokenAmount(credits int64, decimals int) string {
	amount := big.NewInt(credits)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amount.Mul(amount, multiplier)
	return amount.String()
}

// GetTokenDecimals returns the decimals for a token symbol
func GetTokenDecimals(symbol string) int {
	// USDC and USDT both have 6 decimals
	switch symbol {
	case "USDC", "USDT":
		return 6
	default:
		return 18 // Default to 18 for unknown tokens
	}
}

// ChainIDToNetwork maps chain IDs to DFNS network names
var ChainIDToNetwork = map[int64]string{
	1:          "EthereumMainnet",
	11155111:   "EthereumSepolia",
	728126428:  "Tron",
	3448148188: "TronNile",
}

// NetworkToChainID maps DFNS network names to chain IDs
var NetworkToChainID = map[string]int64{
	"EthereumMainnet": 1,
	"EthereumSepolia": 11155111,
	"Tron":            728126428,
	"TronNile":        3448148188,
}

// ChainNameToNetwork maps our chain names to DFNS network names
var ChainNameToNetwork = map[string]string{
	"ethereum":         "EthereumMainnet",
	"ethereum-sepolia": "EthereumSepolia",
	"tron":             "Tron",
	"tron-nile":        "TronNile",
}

// GetDFNSNetwork returns the DFNS network name for a chain name
func GetDFNSNetwork(chainName string) string {
	return ChainNameToNetwork[chainName]
}

// GetChainIDFromNetwork returns the chain ID for a DFNS network name
func GetChainIDFromNetwork(network string) int64 {
	return NetworkToChainID[network]
}
