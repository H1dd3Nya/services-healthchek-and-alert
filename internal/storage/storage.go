package storage

// Интерфейс Storage и общие типы

type Storage interface {
	AddSite(site Site) error
	GetSite(id string) (*Site, error)
	ListSites() ([]Site, error)
	DeleteSite(id string) error
	AddCheckHistory(history SiteCheckHistory) error
	ListCheckHistory(siteID string, limit int) ([]SiteCheckHistory, error)
}
