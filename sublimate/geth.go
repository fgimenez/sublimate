package sublimate

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/release"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
)

const (
	genesisContent = `{
    "nonce": "0x0000000000000042",     "timestamp": "0x0",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "extraData": "0x0",     "gasLimit": "0x8000000",     "difficulty": "0x400",
    "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "coinbase": "0x3333333333333333333333333333333333333333",     "alloc": {     }
}`
)

var (
	relOracle = common.HexToAddress("0xfa7b9770ca4cb04296cac84f37736d4041251cdf")

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

type geth struct{}

type gethConfig struct {
	Eth  eth.Config
	Node node.Config
}

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func (g *geth) run(p *Project) error {
	if err := initTestnet(p); err != nil {
		return err
	}

	// compile contract

	// execute script

	return nil
}

func initTestnet(p *Project) error {
	genesis := new(core.Genesis)

	if err := json.Unmarshal([]byte(genesisContent), genesis); err != nil {
		return errors.New("invalid genesis file: " + err.Error())
	}

	// Open an initialise both full and light databases
	ctx := cli.NewContext(cli.NewApp(), nil, nil)

	stack, err := makeFullNode(ctx)
	if err != nil {
		return err
	}
	for _, name := range []string{"chaindata", "lightchaindata"} {
		chaindb, err := stack.OpenDatabase(name, 0, 0)
		if err != nil {
			utils.Fatalf("Failed to open database: %v", err)
		}
		if _, _, err = core.SetupGenesisBlock(chaindb, genesis); err != nil {
			utils.Fatalf("Failed to write genesis block: %v", err)
		}
	}

	return nil
}

func makeFullNode(ctx *cli.Context) (*node.Node, error) {
	stack, cfg, err := makeConfigNode(ctx)
	if err != nil {
		return nil, err
	}

	utils.RegisterEthService(stack, &cfg.Eth)

	// Add the release oracle service so it boots along with node.
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		config := release.Config{
			Oracle: relOracle,
			Major:  uint32(params.VersionMajor),
			Minor:  uint32(params.VersionMinor),
			Patch:  uint32(params.VersionPatch),
		}
		commit, _ := hex.DecodeString("")
		copy(config.Commit[:], commit)
		return release.NewReleaseService(ctx, config)
	}); err != nil {
		utils.Fatalf("Failed to register the Geth release oracle service: %v", err)
	}
	return stack, nil
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig, error) {
	// Load defaults.
	cfg := gethConfig{
		Eth:  eth.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			return nil, cfg, err
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		return nil, cfg, errors.New("Failed to create the protocol stack: " + err.Error())
	}
	utils.SetEthConfig(ctx, stack, &cfg.Eth)

	return stack, cfg, nil
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = "geth"
	cfg.Version = params.VersionWithCommit("")
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "geth.ipc"
	return cfg
}

func loadConfig(file string, cfg *gethConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}
