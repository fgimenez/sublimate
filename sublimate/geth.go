package sublimate

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
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
    "config": {
        "chainId": 15,
        "homesteadBlock": 0,
        "eip155Block": 0,
        "eip158Block": 0
    },
    "difficulty": "200000000",
    "gasLimit": "2100000",
    "alloc": {
        "7df9a875a174b3bc565e6424a0050ebc1b2d1d82": { "balance": "300000" },
        "f41c74c9ae680c1aa78f42e5647a62f353b7bdde": { "balance": "400000" }
    }
}`
)

var (
	relOracle = common.HexToAddress("0xfa7b9770ca4cb04296cac84f37736d4041251cdf")

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
	datadir string
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
	var err error
	datadir, err = ioutil.TempDir("", "sublimate-datadir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(datadir)

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
	f := flag.NewFlagSet("f1", flag.ContinueOnError)
	ctx := cli.NewContext(cli.NewApp(), f, nil)

	stack, err := makeFullNode(ctx)
	if err != nil {
		return err
	}
	for _, name := range []string{"chaindata", "lightchaindata"} {
		chaindb, err := stack.OpenDatabase(name, 0, 0)
		if err != nil {
			return errors.New("Failed to open database: " + err.Error())
		}
		if _, _, err = core.SetupGenesisBlock(chaindb, genesis); err != nil {
			return errors.New("Failed to write genesis block: " + err.Error())
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
		return nil, errors.New("Failed to register the Geth release oracle service: " + err.Error())
	}
	return stack, nil
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig, error) {
	// Load defaults.
	cfg := gethConfig{
		Eth:  eth.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	cfg.Eth.NetworkId = 15

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
	cfg.DataDir = datadir
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
