package usecase

type DownloadVideosCreator struct {
	// Inject interfaces here (e.g., database or API client)
}

func (d *DownloadVideosCreator) Execute(dst_dir string) error {
	return nil
}
