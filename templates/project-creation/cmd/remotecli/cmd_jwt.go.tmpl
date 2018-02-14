package remotecli

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"
)

func (c *remoteCli) cmdJWT(args []string) (string, error) {
	if len(args) != 0 {
		jwt := args[0]
		if jwt == "" {
			c.UnsetJWT()
		} else {
			c.SetJWT(jwt)
		}
	}

	ctx, ok := metadata.FromOutgoingContext(c.ctx)
	if !ok {
		return "", fmt.Errorf("Context FromOutgoingContext fail")
	}

	v := strings.Join(ctx["authorization"], "")
	if strings.HasPrefix(v, "Bearer ") {
		v = strings.TrimPrefix(v, "Bearer ")
	}

	return fmt.Sprintf("jwt : %s", v), nil
}
