package akka

import (
	"context"
	"encoding/json"
	"fmt"
)

// RegionModel represents a deployment region available in the organization.
type RegionModel struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Provider    string `json:"provider"`
}

func (c *AkkaClient) ListRegions(ctx context.Context) ([]RegionModel, error) {
	out, err := c.Run(ctx, "regions", "list")
	if err != nil {
		return nil, fmt.Errorf("list regions: %w", err)
	}
	var regions []RegionModel
	if err := json.Unmarshal(out, &regions); err != nil {
		var wrapper struct {
			Regions []RegionModel `json:"regions"`
		}
		if err2 := json.Unmarshal(out, &wrapper); err2 != nil {
			return nil, fmt.Errorf("parse regions: %w", err)
		}
		return wrapper.Regions, nil
	}
	return regions, nil
}
