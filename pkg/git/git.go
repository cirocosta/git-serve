package git

import "context"

type git struct {
	dataDir string
}

func (g *git) initRepository(ctx context.Context, name string) (*Repository, error) {
	return nil, nil
}
