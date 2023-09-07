package cache

import (
	"context"
	"log"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

type TokenCache struct {
	File string
}

func (t *TokenCache) Replace(ctx context.Context, cache cache.Unmarshaler, hints cache.ReplaceHints) error {
	data, _ := os.ReadFile(t.File)
	return cache.Unmarshal(data)
}

func (t *TokenCache) Export(ctx context.Context, cache cache.Marshaler, hints cache.ExportHints) error {
	data, err := cache.Marshal()
	if err != nil {
		log.Println(err)
	}
	return os.WriteFile(t.File, data, 0600)
}
