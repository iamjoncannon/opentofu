package plugin6

import (
	"github.com/opentofu/opentofu/internal/tfplugin6"
	"google.golang.org/grpc"
)

func (p *GRPCProvider) GetPlatformConfiguration() {
	const maxRecvSize = 64 << 20

	in := &tfplugin6.GetPlatformConfiguration_Request{}

	getPlatformConfiguration_Response, err := p.configuration_provider_client.GetPlatformConfiguration(p.ctx, in, grpc.MaxRecvMsgSizeCallOption{MaxRecvMsgSize: maxRecvSize})

	if err != nil {
		logger.Trace("GRPCProvider GetPlatformConfiguration", "err", err)
		return
	}

	logger.Trace("GRPCProvider GetPlatformConfiguration getPlatformConfiguration_Response ", "response", getPlatformConfiguration_Response)

}
