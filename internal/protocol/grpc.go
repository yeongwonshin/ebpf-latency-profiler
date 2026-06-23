package protocol

import "strings"

type GRPCHint struct {
	Service string
	Method  string
	Status  int
}

func ParseGRPCMethod(fullMethod string) GRPCHint {
	trimmed := strings.Trim(fullMethod, "/")
	if trimmed == "" {
		return GRPCHint{}
	}
	idx := strings.LastIndex(trimmed, "/")
	if idx < 0 {
		return GRPCHint{Method: trimmed}
	}
	return GRPCHint{Service: trimmed[:idx], Method: trimmed[idx+1:]}
}
