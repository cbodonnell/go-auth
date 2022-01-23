package workers

import (
	"fmt"
	"log"
	"time"

	"github.com/cheebz/go-auth/repositories"
)

type PurgeRefreshWorker struct {
	Repo repositories.Repository
}

func NewPurgeRefreshWorker(repo repositories.Repository) *PurgeRefreshWorker {
	return &PurgeRefreshWorker{
		Repo: repo,
	}
}

// Daily purge of expired refresh tokens
func (w *PurgeRefreshWorker) Start() {
	for {
		err := w.Repo.DeleteExpiredRefresh()
		if err != nil {
			log.Println(fmt.Sprintf("failed to purge refresh: %s", err.Error()))
		}
		time.Sleep(24 * time.Hour)
	}
}
