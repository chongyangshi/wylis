package outgoing

import "context"

func Init(ctx context.Context) error {
	err := initTyphonClient(ctx)
	if err != nil {
		return err
	}

	err = initClusterClient(ctx)
	if err != nil {
		return err
	}

	err = initOutgoing(ctx)
	if err != nil {
		return err
	}

	return nil
}
