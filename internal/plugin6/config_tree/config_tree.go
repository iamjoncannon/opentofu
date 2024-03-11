package config_tree

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/opentofu/opentofu/internal/configs"
	"github.com/opentofu/opentofu/internal/configs/configload"
	"github.com/opentofu/opentofu/internal/logging"
	"github.com/opentofu/opentofu/internal/tfplugin6"
)

var logger = logging.HCLogger()

type ConfigTreeBuilder interface {
	Build() (*configs.Config, *configload.Snapshot)
}

type DemoConfigTreeBuilder struct {
	Tracer                 func(string, interface{})
	Module_shard_container ModuleShardContainer
	RootModuleInstanceNode *ModuleInstanceNode
	ModuleClassMap         ModuleClassMap
	snapshot_map           *configload.Snapshot
}

var Config_tree_builder ConfigTreeBuilder = &DemoConfigTreeBuilder{}

func (c *DemoConfigTreeBuilder) Build() (*configs.Config, *configload.Snapshot) {

	main_module_shards := c.Module_shard_container.ModuleShardMap["main"]
	logger.Debug(fmt.Sprintf("main_module_shards %v", main_module_shards))

	main_module := c.get_module_class_from_shards(main_module_shards)

	logger.Debug(fmt.Sprintf("main_module %v", main_module))

	main_module_instance_node := c.new_main_module_instance_node(*main_module)
	logger.Debug(fmt.Sprintf("main_module_instance_node %v", main_module_instance_node))

	// create snapshot map from shards
	c.snapshot_map = &configload.Snapshot{
		Modules: make(map[ModuleInstanceId]*configload.SnapshotModule),
	}

	c.RootModuleInstanceNode = &main_module_instance_node

	c.ModuleClassMap = make(ModuleClassMap)
	c.snapshot_map.Modules["main"] = c.get_module_snapshot([]string{"."}, "main", main_module_shards)

	c.walk_node(&main_module_instance_node, nil)

	return c.RootModuleInstanceNode, c.snapshot_map
}

func (c *DemoConfigTreeBuilder) walk_node(module_instance_node *ModuleInstanceNode, parent_instance_node *ModuleInstanceNode) {

	// populate parent

	module_instance_node.Parent = parent_instance_node

	// populate children

	module_calls := module_instance_node.Module.ModuleCalls

	for _, module_call := range module_calls {

		child_module_instance_id := module_call.Name
		child_module_class_id := c.Module_shard_container.ModuleInstanceToClassMap[child_module_instance_id]

		// get module class

		child_module_class, is_already_created := c.ModuleClassMap[child_module_class_id]
		module_class_shards := c.Module_shard_container.ModuleShardMap[child_module_class_id]

		if !is_already_created {
			child_module_class = c.get_module_class_from_shards(module_class_shards)
			c.ModuleClassMap[child_module_class_id] = child_module_class
		}

		module_call_range := module_call.SourceAddrRange

		module_instance_path := []string{child_module_instance_id}

		child_module_instance := c.new_module_instance_node(
			child_module_class,
			module_instance_path, // todo- add parent module instance path
			module_call_range,
		)

		// add instance to parent children map
		module_instance_node.Children[child_module_instance_id] = child_module_instance

		// add entry to snapshot map
		c.snapshot_map.Modules[child_module_instance_id] = c.get_module_snapshot(module_instance_path, child_module_instance_id, module_class_shards)

		if !is_already_created {
			c.walk_node(child_module_instance, module_instance_node)
		}

	}

}

func (c *DemoConfigTreeBuilder) get_module_class_from_shards(module_shards []*tfplugin6.ModuleShard) *configs.Module {
	parser := hclparse.NewParser()

	tf_files := []*configs.File{}

	for _, shard := range module_shards {

		hclFile, parseDiags := parser.ParseHCL(shard.RawFileContainer.Bytes, shard.RawFileContainer.Path)

		if parseDiags.HasErrors() {
			logger.Debug("parseDiags ", parseDiags.Errs())
		}

		tf_config_file := &configs.File{}

		content, _, contentDiags := hclFile.Body.PartialContent(TerraformSchema)

		if contentDiags.HasErrors() {
			logger.Debug("contentDiags ", parseDiags.Errs())
		}

		for _, superficial_block := range content.Blocks {
			fileDiags := hcl.Diagnostics{}
			configs.DecodeBlock(superficial_block, fileDiags, tf_config_file, false)
			logger.Debug(fmt.Sprintf("fileDiags %v ", fileDiags))
		}

		tf_files = append(tf_files, tf_config_file)
	}

	module_class, err := configs.NewModule(tf_files, nil)

	if err != nil {
		logger.Debug("module err", err)
	}

	return module_class
}

var TerraformSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "terraform",
			LabelNames: nil,
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
		{
			Type:       "output",
			LabelNames: []string{"name"},
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
	},
}
