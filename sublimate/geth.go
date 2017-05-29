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
	"os/signal"
	"reflect"
	"strings"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console"
	"github.com/ethereum/go-ethereum/contracts/release"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
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

	if err := compileContract(p.Contract); err != nil {
		return err
	}

	if err := executeScript(p.Script); err != nil {
		return err
	}

	return nil
}

func initTestnet(p *Project) error {
	genesis := new(core.Genesis)

	if err := json.Unmarshal([]byte(genesisContent), genesis); err != nil {
		return errors.New("invalid genesis file: " + err.Error())
	}

	// Open and initialise both full and light databases
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

func compileContract(contract string) error {
	_, err := ioutil.ReadFile(contract)
	if err != nil {
		return err
	}

	script := ``

	return executeScript(script)
}

func executeScript(script string) error {
	// create script file

	// Create and start the node based on the CLI flags
	f := flag.NewFlagSet("f1", flag.ContinueOnError)
	ctx := cli.NewContext(cli.NewApp(), f, nil)

	node, err := makeFullNode(ctx)
	if err != nil {
		return err
	}
	startNode(ctx, node)
	defer node.Stop()

	// Attach to the newly started node and start the JavaScript console
	client, err := node.Attach()
	if err != nil {
		return errors.New("Failed to attach to the inproc geth: " + err.Error())
	}
	config := console.Config{
		DataDir: datadir,
		DocRoot: ctx.GlobalString(utils.JSpathFlag.Name),
		Client:  client,
		Preload: utils.MakeConsolePreloads(ctx),
	}

	console, err := console.New(config)
	if err != nil {
		return errors.New("Failed to start the JavaScript console: " + err.Error())
	}
	defer console.Stop(false)

	// Evaluate each of the specified JavaScript files
	scriptfile, err := ioutil.TempFile("", "eth-script")
	if err != nil {
		return err
	}

	defer os.Remove(scriptfile.Name())

	if _, err := scriptfile.Write([]byte(script)); err != nil {
		return err
	}
	if err := scriptfile.Close(); err != nil {
		return err
	}
	if err = console.Execute(scriptfile.Name()); err != nil {
		return errors.New("Failed to execute: " + err.Error())
	}
	// Wait for pending callbacks, but stop for Ctrl-C.
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, os.Interrupt)

	go func() {
		<-abort
		os.Exit(0)
	}()
	console.Stop(true)

	return nil
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node) {
	// Start up the node itself
	utils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	passwords := utils.MakePasswordList(ctx)
	unlocks := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for i, account := range unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			unlockAccount(ctx, ks, trimmed, i, passwords)
		}
	}
	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	go func() {
		// Create an chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			utils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := ethclient.NewClient(rpcClient)

		// Open and self derive any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			} else {
				wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			if event.Arrive {
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				} else {
					log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", event.Wallet.Status())
					event.Wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
				}
			} else {
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()
}

// tries unlocking the specified account a few times.
func unlockAccount(ctx *cli.Context, ks *keystore.KeyStore, address string, i int, passwords []string) (accounts.Account, string) {
	account, err := utils.MakeAddress(ks, address)
	if err != nil {
		utils.Fatalf("Could not list accounts: %v", err)
	}
	for trials := 0; trials < 3; trials++ {
		prompt := fmt.Sprintf("Unlocking account %s | Attempt %d/%d", address, trials+1, 3)
		password := getPassPhrase(prompt, false, i, passwords)
		err = ks.Unlock(account, password)
		if err == nil {
			log.Info("Unlocked account", "address", account.Address.Hex())
			return account, password
		}
		if err, ok := err.(*keystore.AmbiguousAddrError); ok {
			log.Info("Unlocked account", "address", account.Address.Hex())
			return ambiguousAddrRecovery(ks, err, password), password
		}
		if err != keystore.ErrDecrypt {
			// No need to prompt again if the error is not decryption-related.
			break
		}
	}
	// All trials expended to unlock account, bail out
	utils.Fatalf("Failed to unlock account %s (%v)", address, err)

	return accounts.Account{}, ""
}

// getPassPhrase retrieves the passwor associated with an account, either fetched
// from a list of preloaded passphrases, or requested interactively from the user.
func getPassPhrase(prompt string, confirmation bool, i int, passwords []string) string {
	// If a list of passwords was supplied, retrieve from them
	if len(passwords) > 0 {
		if i < len(passwords) {
			return passwords[i]
		}
		return passwords[len(passwords)-1]
	}
	// Otherwise prompt the user for the password
	if prompt != "" {
		fmt.Println(prompt)
	}
	password, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		utils.Fatalf("Failed to read passphrase: %v", err)
	}
	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			utils.Fatalf("Failed to read passphrase confirmation: %v", err)
		}
		if password != confirm {
			utils.Fatalf("Passphrases do not match")
		}
	}
	return password
}

func ambiguousAddrRecovery(ks *keystore.KeyStore, err *keystore.AmbiguousAddrError, auth string) accounts.Account {
	fmt.Printf("Multiple key files exist for address %x:\n", err.Addr)
	for _, a := range err.Matches {
		fmt.Println("  ", a.URL)
	}
	fmt.Println("Testing your passphrase against all of them...")
	var match *accounts.Account
	for _, a := range err.Matches {
		if err := ks.Unlock(a, auth); err == nil {
			match = &a
			break
		}
	}
	if match == nil {
		utils.Fatalf("None of the listed files could be unlocked.")
	}
	fmt.Printf("Your passphrase unlocked %s\n", match.URL)
	fmt.Println("In order to avoid this warning, you need to remove the following duplicate key files:")
	for _, a := range err.Matches {
		if a != *match {
			fmt.Println("  ", a.URL)
		}
	}
	return *match
}
