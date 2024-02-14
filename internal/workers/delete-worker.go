package workers

import (
	"context"

	"github.com/ma-shulgin/go-link-shortener/internal/logger"
	"github.com/ma-shulgin/go-link-shortener/internal/storage"
)

type DeleteContext struct {
	ShortURLs []string
	Ctx       context.Context
}

func RunDeleteWorker(urlStore storage.URLStore, chanSize int) chan DeleteContext {
	deleteCh := make(chan DeleteContext, chanSize)
	go func() {
		for data := range deleteCh {
			if err := urlStore.DeleteURLs(data.Ctx, data.ShortURLs); err != nil {
				logger.Log.Errorln("Failed to delete URLs", err)
			}
		}
	}()
	return deleteCh
}
