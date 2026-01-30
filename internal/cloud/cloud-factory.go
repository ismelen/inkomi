package cloud

import "fmt"

func GetCloud(cloud string) (CloudI, error) {
	switch cloud {
	case "google":
		return &GoogleCloud{}, nil
	case "local": 
		return &LocalCloud{}, nil
	default:
		return nil, fmt.Errorf("No cloud method available: %s", cloud)
	}
}
