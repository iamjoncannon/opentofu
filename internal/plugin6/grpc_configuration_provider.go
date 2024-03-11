package plugin6

import (
	"fmt"
	"io"

	"github.com/opentofu/opentofu/internal/configs"
	"github.com/opentofu/opentofu/internal/configs/configload"
	"github.com/opentofu/opentofu/internal/plugin6/config_tree"
	"github.com/opentofu/opentofu/internal/tfplugin6"
	"google.golang.org/grpc"
)

// WORKING HERE \\
func (p *GRPCProvider) GetPlatformConfiguration() (config *configs.Config, configSnap *configload.Snapshot) {

	const maxRecvSize = 64 << 20

	in := &tfplugin6.GetPlatformConfiguration_Request{}

	stream_client, err := p.configuration_provider_client.GetPlatformConfiguration(p.ctx, in, grpc.MaxRecvMsgSizeCallOption{MaxRecvMsgSize: maxRecvSize})

	if err != nil {
		logger.Trace("[ERROR] GRPCProvider GetPlatformConfiguration", "err", err)
		return nil, nil
	}

	module_shard_map := make(config_tree.ModuleShardMap)
	module_instance_to_class_map := make(config_tree.ModuleInstanceToClassMap)

	for {

		response, err := stream_client.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			logger.Trace(fmt.Sprintf("err module_shard %+v %+v", response, err))
			break
		}

		if response.ModuleShard != nil {
			module_class_id := response.ModuleShard.ModuleClassId
			module_shard_map[module_class_id] = append(module_shard_map[module_class_id], response.ModuleShard)
		}

		if response.ModuleInstanceToClassMap != nil && len(response.ModuleInstanceToClassMap) > 0 {
			module_instance_to_class_map = response.ModuleInstanceToClassMap
		}

	}

	module_shard_container := config_tree.ModuleShardContainer{
		ModuleShardMap:           module_shard_map,
		ModuleInstanceToClassMap: module_instance_to_class_map,
	}

	config_tree_builder := config_tree.DemoConfigTreeBuilder{
		Module_shard_container: module_shard_container,
	}

	return config_tree_builder.Build()
}

/*

modules.json in .terraform/modules folder

"Source"- for local, this is the relative file path
"Dir"- for local, this is the subdirectory in the modules folder without the relative file path prefix

{
	"Modules":[
		{"Key":"","Source":"","Dir":"."},  													// main module

		{"Key":"network_east","Source":"./modules/networking","Dir":"modules/networking"},

		{"Key":"network_west","Source":"./modules/networking","Dir":"modules/networking"}
	]
}


*/

/*

structure of normal Config Tree

in our conceptual schema
	- "Config" maps to the module instance tree structure
	- the Module property is the module class for each instance

main Config struct

config &{
	Root:0xc00063c0e0 				// pointing to itself
	Parent:<nil>
	Path:
	Children:map[
		network_east:0xc00063d500	// key is module instance name, value is Config struct
		network_west:0xc00063d6c0
	]
	Module:0xc00063e270
	CallRange::0,0-0 				// main config struct has no call range, so call range must refer to the ModuleCall code location
	SourceAddr:<nil>
	SourceAddrRaw:
	SourceAddrRange::0,0-0 			// addr range for main is also null
	Version:<nil>
}

Module &{
	SourceDir:.
	CoreVersionConstraints:[]
	ActiveExperiments:map[]
	Backend:<nil>
	CloudConfig:<nil>
	ProviderConfigs:map[aws.east:0xc0006faf00 aws.west:0xc0006fb040]
	ProviderRequirements:0xc000503ef0
	ProviderLocalNames:map[registry.opentofu.org/hashicorp/aws:aws]
	ProviderMetas:map[]
	Variables:map[private_subnet_cidrs:0xc0008aab60 public_subnet_cidrs:0xc0008aa9c0]
	Locals:map[]
	Outputs:map[]
	ModuleCalls:map[network_east:0xc00081b680 network_west:0xc00081b500]
	ManagedResources:map[]
	DataResources:map[]
	Moved:[]
	Import:[]
	Checks:map[]
	Tests:map[]
}


===============

module.network_east Config struct

config &{
	Root:0xc000829c00
	Parent:0xc000829c00
	Path:module.network_east
	Children:map[]
	Module:0xc0008e75f0
	CallRange:main.tf:39,1-22 				// code location of the module instance name label
	SourceAddr:./modules/networking
	SourceAddrRaw:
	SourceAddrRange:main.tf:40,12-34		// code location of the source directory property
	Version:<nil>
}

Module &{
	SourceDir:modules/networking
	CoreVersionConstraints:[]
	ActiveExperiments:map[]
	Backend:<nil>
	CloudConfig:<nil>
	ProviderConfigs:map[]
	ProviderRequirements:0xc0002d4d70
	ProviderLocalNames:map[registry.opentofu.org/hashicorp/aws:aws]
	ProviderMetas:map[]
	Variables:map[base_cidr_block:0xc0008e7040 private_subnet_cidrs:0xc0008e72b0 public_subnet_cidrs:0xc0008e71e0 region:0xc0008e7110]
	Locals:map[] Outputs:map[] ModuleCalls:map[]
	ManagedResources:map[aws_internet_gateway.main:0xc0006f5860 aws_subnet.private-a:0xc0006f51e0 aws_subnet.private-b:0xc0006f5520 aws_subnet.public-a:0xc0006f5040 aws_subnet.public-b:0xc0006f5380 aws_vpc.main:0xc0006f56c0]
	DataResources:map[]
	Moved:[]
	Import:[]
	Checks:map[]
	Tests:map[]
}

*/

/*

configSnap &{
	Modules:map[
		:0xc0008df560 					// main module's module path is ""
		network_east:0xc0008a3e30 		// key is module instance name
		network_west:0xc0008def60
		]
	}


snapshotModule for main

&{
	Dir:. 								// directory for main is "."

	Files:map[
		globals.tf:[10 118 97 114 105 97 98 108 101 32 34 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 34 32 123 10 32 32 100 101 102 97 117 108 116 32 61 32 91 34 49 48 46 48 46 49 46 48 47 50 52 34 44 32 34 49 48 46 48 46 50 46 48 47 50 52 34 44 32 34 49 48 46 48 46 51 46 48 47 50 52 34 93 10 125 10 10 118 97 114 105 97 98 108 101 32 34 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 34 32 123 10 32 32 100 101 102 97 117 108 116 32 61 32 91 34 49 48 46 48 46 52 46 48 47 50 52 34 44 32 34 49 48 46 48 46 53 46 48 47 50 52 34 44 32 34 49 48 46 48 46 54 46 48 47 50 52 34 93 10 125 10] main.tf:[116 101 114 114 97 102 111 114 109 32 123 10 32 32 114 101 113 117 105 114 101 100 95 112 114 111 118 105 100 101 114 115 32 123 10 10 32 32 32 32 97 119 115 32 61 32 123 10 32 32 32 32 32 32 115 111 117 114 99 101 32 32 61 32 34 104 97 115 104 105 99 111 114 112 47 97 119 115 34 10 32 32 32 32 32 32 118 101 114 115 105 111 110 32 61 32 34 51 46 55 52 46 48 34 10 32 32 32 32 125 10 10 32 32 125 10 125 10 10 10 109 111 100 117 108 101 32 34 110 101 116 119 111 114 107 95 119 101 115 116 34 32 123 10 32 32 115 111 117 114 99 101 32 61 32 34 46 47 109 111 100 117 108 101 115 47 110 101 116 119 111 114 107 105 110 103 34 10 32 32 98 97 115 101 95 99 105 100 114 95 98 108 111 99 107 32 61 32 34 49 48 46 48 46 48 46 48 47 50 50 34 10 32 32 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 32 61 32 118 97 114 46 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 10 32 32 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 32 61 32 118 97 114 46 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 10 32 32 114 101 103 105 111 110 32 61 32 34 117 115 45 119 101 115 116 45 49 34 10 10 32 32 112 114 111 118 105 100 101 114 115 32 61 32 123 10 32 32 32 32 97 119 115 32 61 32 97 119 115 46 119 101 115 116 10 32 32 125 10 125 10 10 109 111 100 117 108 101 32 34 110 101 116 119 111 114 107 95 101 97 115 116 34 32 123 10 32 32 115 111 117 114 99 101 32 61 32 34 46 47 109 111 100 117 108 101 115 47 110 101 116 119 111 114 107 105 110 103 34 10 32 32 98 97 115 101 95 99 105 100 114 95 98 108 111 99 107 32 61 32 34 49 48 46 48 46 48 46 48 47 50 50 34 10 32 32 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 32 61 32 118 97 114 46 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 10 32 32 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 32 61 32 118 97 114 46 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 10 32 32 114 101 103 105 111 110 32 61 32 34 117 115 45 101 97 115 116 45 49 34 10 32 32 112 114 111 118 105 100 101 114 115 32 61 32 123 10 32 32 32 32 97 119 115 32 61 32 97 119 115 46 101 97 115 116 10 32 32 125 10 125 10]
		providers.tf:[112 114 111 118 105 100 101 114 32 34 97 119 115 34 32 123 10 32 32 97 108 105 97 115 32 32 32 32 32 61 32 34 101 97 115 116 34 10 32 32 114 101 103 105 111 110 32 32 32 32 32 61 32 34 117 115 45 101 97 115 116 45 49 34 10 32 32 97 99 99 101 115 115 95 107 101 121 32 61 32 34 65 75 73 65 85 69 66 78 53 68 86 84 78 51 84 84 67 51 85 52 34 10 32 32 115 101 99 114 101 116 95 107 101 121 32 61 32 34 99 86 114 57 47 119 88 82 112 107 68 47 116 43 56 72 54 50 57 115 102 71 69 117 105 74 101 113 80 54 113 67 100 109 117 81 74 85 76 117 34 10 125 10 10 112 114 111 118 105 100 101 114 32 34 97 119 115 34 32 123 10 32 32 97 108 105 97 115 32 32 32 32 32 61 32 34 119 101 115 116 34 10 32 32 114 101 103 105 111 110 32 32 32 32 32 61 32 34 117 115 45 119 101 115 116 45 49 34 10 32 32 97 99 99 101 115 115 95 107 101 121 32 61 32 34 65 75 73 65 85 69 66 78 53 68 86 84 78 51 84 84 67 51 85 52 34 10 32 32 115 101 99 114 101 116 95 107 101 121 32 61 32 34 99 86 114 57 47 119 88 82 112 107 68 47 116 43 56 72 54 50 57 115 102 71 69 117 105 74 101 113 80 54 113 67 100 109 117 81 74 85 76 117 34 10 125]]

	SourceAddr:							// SourceAddress is ""

	Version:<nil>
}


snapshotModule network_east and network_west

&{
	Dir:modules/networking 					// Dir is the SourceAddr without the relative file path at the start "./"

	Files:map[
		subnets.tf:[ 46 105 100 10 32 32 99 105 100 114 95 98 108 111 99 107 32 32 32 32 32 32 32 32 32 32 32 32 32 32 61 32 118 97 114 46 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 91 48 93 10 32 32 97 118 97 105 108 97 98 105 108 105 116 121 95 122 111 110 101 32 32 32 32 32 32 32 61 32 34 36 123 118 97 114 46 114 101 103 105 111 110 125 97 34 10 32 32 109 97 112 95 112 117 98 108 105 99 95 105 112 95 111 110 95 108 97 117 110 99 104 32 61 32 116 114 117 101 10 10 32 32 116 97 103 115 32 61 32 123 10 32 32 32 32 34 78 97 109 101 34 32 61 32 34 112 117 98 108 105 99 45 36 123 118 97 114 46 114 101 103 105 111 110 125 97 34 10 32 32 125 10 125 10 10 114 101 115 111 117 114 99 101 32 34 97 119 115 95 115 117 98 110 101 116 34 32 34 112 114 105 118 97 116 101 45 97 34 32 123 10 32 32 118 112 99 95 105 100 32 32 32 32 32 61 32 97 119 115 95 118 112 99 46 109 97 105 110 46 105 100 10 32 32 99 105 100 114 95 98 108 111 99 107 32 61 32 118 97 114 46 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 91 48 93 10 10 32 32 97 118 97 105 108 97 98 105 108 105 116 121 95 122 111 110 101 32 61 32 34 36 123 118 97 114 46 114 101 103 105 111 110 125 97 34 10 10 32 32 116 97 103 115 32 61 32 123 10 32 32 32 32 34 78 97 109 101 34 32 61 32 34 112 114 105 118 97 116 101 45 36 123 118 97 114 46 114 101 103 105 111 110 125 97 34 10 32 32 125 10 125 10 10 114 101 115 111 117 114 99 101 32 34 97 119 115 95 115 117 98 110 101 116 34 32 34 112 117 98 108 105 99 45 98 34 32 123 10 32 32 118 112 99 95 105 100 32 32 32 32 32 32 32 32 32 32 32 32 32 32 32 32 32 32 61 32 97 119 115 95 118 112 99 46 109 97 105 110 46 105 100 10 32 32 99 105 100 114 95 98 108 111 99 107 32 32 32 32 32 32 32 32 32 32 32 32 32 32 61 32 118 97 114 46 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 91 49 93 10 32 32 97 118 97 105 108 97 98 105 108 105 116 121 95 122 111 110 101 32 32 32 32 32 32 32 61 32 34 36 123 118 97 114 46 114 101 103 105 111 110 125 98 34 10 32 32 109 97 112 95 112 117 98 108 105 99 95 105 112 95 111 110 95 108 97 117 110 99 104 32 61 32 116 114 117 101 10 10 32 32 116 97 103 115 32 61 32 123 10 32 32 32 32 34 78 97 109 101 34 32 61 32 34 112 117 98 108 105 99 45 36 123 118 97 114 46 114 101 103 105 111 110 125 98 34 10 32 32 125 10 125 10 10 114 101 115 111 117 114 99 101 32 34 97 119 115 95 115 117 98 110 101 116 34 32 34 112 114 105 118 97 116 101 45 98 34 32 123 10 32 32 118 112 99 95 105 100 32 32 32 32 32 32 32 32 32 32 32 32 61 32 97 119 115 95 118 112 99 46 109 97 105 110 46 105 100 10 32 32 99 105 100 114 95 98 108 111 99 107 32 32 32 32 32 32 32 32 61 32 118 97 114 46 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 91 49 93 10 32 32 97 118 97 105 108 97 98 105 108 105 116 121 95 122 111 110 101 32 61 32 34 36 123 118 97 114 46 114 101 103 105 111 110 125 98 34 10 10 32 32 116 97 103 115 32 61 32 123 10 32 32 32 32 34 78 97 109 101 34 32 61 32 34 112 114 105 118 97 116 101 45 36 123 118 97 114 46 114 101 103 105 111 110 125 98 34 10 32 32 125 10 125 10]
		vpc.tf:[10 118 97 114 105 97 98 108 101 32 34 98 97 115 101 95 99 105 100 114 95 98 108 111 99 107 34 32 123 125 10 118 97 114 105 97 98 108 101 32 34 112 117 98 108 105 99 95 115 117 98 110 101 116 95 99 105 100 114 115 34 32 123 125 10 118 97 114 105 97 98 108 101 32 34 112 114 105 118 97 116 101 95 115 117 98 110 101 116 95 99 105 100 114 115 34 32 123 125 10 118 97 114 105 97 98 108 101 32 34 114 101 103 105 111 110 34 32 123 125 10 10 116 101 114 114 97 102 111 114 109 32 123 10 32 32 114 101 113 117 105 114 101 100 95 112 114 111 118 105 100 101 114 115 32 123 10 10 32 32 32 32 97 119 115 32 61 32 123 10 32 32 32 32 32 32 115 111 117 114 99 101 32 32 61 32 34 104 97 115 104 105 99 111 114 112 47 97 119 115 34 10 32 32 32 32 32 32 118 101 114 115 105 111 110 32 61 32 34 51 46 55 52 46 48 34 10 32 32 32 32 125 10 10 32 32 125 10 125 10 10 114 101 115 111 117 114 99 101 32 34 97 119 115 95 118 112 99 34 32 34 109 97 105 110 34 32 123 10 32 32 99 105 100 114 95 98 108 111 99 107 32 61 32 118 97 114 46 98 97 115 101 95 99 105 100 114 95 98 108 111 99 107 10 125 10 10 114 101 115 111 117 114 99 101 32 34 97 119 115 95 105 110 116 101 114 110 101 116 95 103 97 116 101 119 97 121 34 32 34 109 97 105 110 34 32 123 10 32 32 118 112 99 95 105 100 32 61 32 97 119 115 95 118 112 99 46 109 97 105 110 46 105 100 10 125 10 10]
		]
	SourceAddr:./modules/networking 		// relative path of module path

	Version:<nil>
	}

*/
