package config_tree

import (
	"github.com/opentofu/opentofu/internal/configs"
	config_load "github.com/opentofu/opentofu/internal/configs/configload"
	"github.com/opentofu/opentofu/internal/tfplugin6"

	tf_core_config "github.com/opentofu/opentofu/internal/configs"
)

type ModuleInstanceNode = configs.Config
type ModuleClassId = string
type ModuleInstanceId = string

// a shard of the total module fileset
// we shard to stream large modules
// efficiently given the serialization
// issues with the current tf structs
// type ModuleShard struct {
// 	ModuleClassId    string
// 	Shard            int
// 	TotalShards      int
// 	RawFileContainer *RawFileContainer
// }

type ModuleShardMap = map[ModuleClassId][]*tfplugin6.ModuleShard
type ModuleInstanceToClassMap = map[ModuleInstanceId]ModuleClassId

type ModuleClassMap = map[ModuleClassId]*tf_core_config.Module

type ModuleShardContainer struct {
	ModuleShardMap ModuleShardMap

	// translate the 'edges' in the module instance tree
	// for the host--  the interpreter can control the relationship and
	// decouple it from the default tf module implementation
	// (e.g. folder structure, etc)
	ModuleInstanceToClassMap ModuleInstanceToClassMap
}

type RawFileContainer struct {
	Bytes []byte
	Path  string
}

func (r RawFileContainer) get_file_name_from_path() {}

// the value seems to refer to the ModuleClass, but the snapshot id
// is the instance name, e.g. "network_east"
type ModuleSnapshot = map[ModuleInstanceId]config_load.SnapshotModule
