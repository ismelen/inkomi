package cloud

type CloudflareR2Storage struct{}

func NewCloudflareR2Storage() *CloudflareR2Storage { return &CloudflareR2Storage{} }

func (c *CloudflareR2Storage) GetUrl(id string, ext string) (string, error) {
	return "", nil
}
