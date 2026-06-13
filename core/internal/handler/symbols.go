package handler

import (
	"strings"

	"stock-market/backend/internal/service"
)

const maxWatchlistSymbols = 50

func parseSymbolsQuery(symbolParam, symbolsParam string) ([]string, error) {
	var raw []string

	if strings.TrimSpace(symbolsParam) != "" {
		raw = strings.Split(symbolsParam, ",")
	} else if strings.TrimSpace(symbolParam) != "" {
		raw = []string{symbolParam}
	}

	if len(raw) == 0 {
		return nil, service.ErrMissingSymbols
	}

	seen := make(map[string]struct{}, len(raw))
	symbols := make([]string, 0, len(raw))

	for _, item := range raw {
		symbol := strings.ToUpper(strings.TrimSpace(item))
		if symbol == "" {
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return nil, service.ErrMissingSymbols
	}

	if len(symbols) > maxWatchlistSymbols {
		return nil, service.ErrTooManySymbols
	}

	return symbols, nil
}
