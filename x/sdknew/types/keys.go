package types

const (
	// ModuleName defines the module name
	ModuleName = "sdknew"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_sdknew"
)

var (
	ParamsKey = []byte("p_sdknew")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
