package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/naoina/toml"

	"github.com/urfave/cli/v2"
)

type Configurations struct {
	Title    string
	Settings struct {
		Moniker          string
		ChainId          string
		Github           string
		Genesis          string
		Daemon           string
		OperatorKey      string
		CommissionRate   string
		ValidatorWebsite string
		Description      string
		FolderName       string
		Home             string
	}

	Sync struct {
		Height int64
		Hash   string
		Rpc    string
	}
	Config struct {
		Seeds           string
		PersistentPeers string
	}
}

func main() {
	var conf = Configurations{Title: "Setting"}
	cmd := exec.Command("/bin/sh", "-c", "mkdir -p ~/.cosmo-nodes")
	cmd2 := exec.Command("/bin/sh", "-c", "cp ~/cosmonodes-manager/config.toml ~/.cosmo-nodes/config.toml")
	err := cmd2.Run()
	if err != nil {
		fmt.Println(err)
	}
	cf, err := os.Open("~/.cosmo-nodes/config.toml")
	if err != nil {
		// failed to create/open the file
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
		}

	}
	if err := toml.NewDecoder(cf).Decode(&conf); err != nil {
		// failed to encode
		fmt.Println("Not decoded")
		log.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		// failed to close the file
		log.Fatal(err)

	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "utils",
				Aliases: []string{"u"},
				Usage:   "helper",
				Subcommands: []*cli.Command{
					{
						Name:    "sync-progress",
						Aliases: []string{"sp"},
						Usage:   "See the sync progress",
						Action: func(c *cli.Context) error {
							resp, err := http.Get("localhost:26657/status")
							if err != nil {
								// handle err
								fmt.Printf("%s", err)
							}
							//We Read the response body on the line below.
							if err != nil {
								log.Fatalln(err)
							}

							var result map[string]interface{}

							json.NewDecoder(resp.Body).Decode(&result)
							height := result["sync_info"].(map[string]interface{})["latest_block_height"]
							fmt.Printf("%s", height)
							return nil
						},
					},
					{
						Name:    "peers",
						Aliases: []string{"sp"},
						Usage:   "See the peers",
						Action: func(c *cli.Context) error {
							resp, err := http.Get("http://localhost:26657/net_info")
							if err != nil {
								// handle err
								fmt.Printf("%s", err)
							}
							//We Read the response body on the line below.
							if err != nil {
								log.Fatalln(err)
							}

							var result map[string]interface{}

							json.NewDecoder(resp.Body).Decode(&result)
							peers := result["peers"]
							fmt.Printf("%s", peers)
							return nil
						},
					},
				},
			},
			{
				Name:    "sync",
				Aliases: []string{"s"},
				Usage:   "Prepare the sync",
				Action: func(c *cli.Context) error {

					// curl -s $SNAP_RPC/block | jq -r .result.block.header.height)

					resp, err := http.Get(conf.Sync.Rpc + "/block")
					if err != nil {
						// handle err
						fmt.Printf("%s", err)
					}
					//We Read the response body on the line below.
					if err != nil {
						log.Fatalln(err)
					}

					var result map[string]interface{}

					json.NewDecoder(resp.Body).Decode(&result)
					height := result["result"].(map[string]interface{})["block"].(map[string]interface{})["header"].(map[string]interface{})["height"]
					var new int
					if s, err := strconv.Atoi(height.(string)); err == nil {
						new = s - 1000
					}

					resp1, err := http.Get(conf.Sync.Rpc + "/block?height=" + strconv.Itoa(new))
					if err != nil {
						log.Fatalln(err)
					}

					var result2 map[string]interface{}

					json.NewDecoder(resp1.Body).Decode(&result2)
					hash := result2["result"].(map[string]interface{})["block_id"].(map[string]interface{})["hash"]
					fmt.Println(`
[statesync]
# State sync rapidly bootstraps a new node by discovering, fetching, and restoring a state machine
# snapshot from peers instead of fetching and replaying historical blocks. Requires some peers in
# the network to take and serve state machine snapshots. State sync is not attempted if the node
# has any local state (LastBlockHeight > 0). The node will have a truncated block history,
# starting from the height of the snapshot.
enable = true

# RPC servers (comma-separated) for light client verification of the synced state machine and
# retrieval of state data for node bootstrapping. Also needs a trusted height and corresponding
# header hash obtained from a trusted source, and a period during which validators can be trusted.
#
# For Cosmos SDK-based chains, trust_period should usually be about 2/3 of the unbonding time (~2
# weeks) during which they can be financially punished (slashed) for misbehavior.`)
					fmt.Println(`trust_height = "` + height.(string) + `"`)
					fmt.Println(`trust_hash = "` + hash.(string) + `"`)
					fmt.Println(`rpc_servers = "` + conf.Sync.Rpc + "," + conf.Sync.Rpc + `"`)
					fmt.Println(`trust_period = "168h0m0s"`)

					return nil
				},
			},
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install it up",
				Action: func(c *cli.Context) error {
					cmd := exec.Command("/bin/sh", "-c", "git clone "+conf.Settings.Github+"; cd "+conf.Settings.FolderName+"; make install; wget -O ~/."+conf.Settings.Home+"/config/genesis.json "+conf.Settings.Genesis)
					err := cmd.Run()
					if err != nil {
						fmt.Println(err)
					}

					return nil
				},
			},

			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configure the validator",
				Flags:   []cli.Flag{},
				Action: func(c *cli.Context) error {
					fmt.Println("Configuring validator...")
					fmt.Println("What is the name of the validator")
					scanner := bufio.NewScanner(os.Stdin)
					scanner.Scan()
					conf.Settings.Moniker = scanner.Text()
					fmt.Println("Hi you, " + conf.Settings.Moniker)

					fmt.Println("On which chain name are you trying to connect?")
					scanner1 := bufio.NewScanner(os.Stdin)
					scanner1.Scan()
					conf.Settings.ChainId = scanner1.Text()
					fmt.Println("We will then connect on: " + conf.Settings.ChainId)
					fmt.Println("The chain has been received.")

					fmt.Println("What is the address of the repo?")
					scanner2 := bufio.NewScanner(os.Stdin)
					scanner2.Scan()
					conf.Settings.Github = scanner2.Text()
					fmt.Println("We will try to download the github on " + conf.Settings.Github)
					fmt.Println("Confirmed we now have the repo's address.")

					fmt.Println("What is the URI to get the genesis file?")
					scanner3 := bufio.NewScanner(os.Stdin)
					scanner3.Scan()
					conf.Settings.Genesis = scanner3.Text()
					fmt.Println("We will get the genesis file from here: " + conf.Settings.Genesis)

					fmt.Println("What is the nme of the daemon?")
					scanner4 := bufio.NewScanner(os.Stdin)
					scanner4.Scan()
					conf.Settings.Daemon = scanner4.Text()
					fmt.Println("All commands will be run with: " + conf.Settings.Daemon)

					fmt.Println("What is the name of the Operator?")
					scanner5 := bufio.NewScanner(os.Stdin)
					scanner5.Scan()
					conf.Settings.OperatorKey = scanner5.Text()
					fmt.Println("All commands will be run with: " + conf.Settings.OperatorKey)

					fmt.Println("What is the commission rate?")
					scanner6 := bufio.NewScanner(os.Stdin)
					scanner6.Scan()
					conf.Settings.CommissionRate = scanner6.Text()
					fmt.Println("Your validator will take a commission of : " + conf.Settings.CommissionRate)

					fmt.Println("What is your website ?")
					scanner7 := bufio.NewScanner(os.Stdin)
					scanner7.Scan()
					conf.Settings.ValidatorWebsite = scanner7.Text()
					fmt.Println("Website is: " + conf.Settings.ValidatorWebsite)

					fmt.Println("What is your description ?")
					scanner8 := bufio.NewScanner(os.Stdin)
					scanner8.Scan()
					conf.Settings.Description = scanner8.Text()
					fmt.Println("Description: " + conf.Settings.Description)
					cmd := exec.Command("/bin/sh", "-c", "mkdir ~/.cosmo-nodes;")
					err := cmd.Run()
					if err != nil {
						fmt.Println(err)
					}

					f, err := os.Create("~/.cosmo-nodes/config.toml")
					if err != nil {
						// failed to create/open the file
						log.Fatal(err)
					}
					if err := toml.NewEncoder(f).Encode(conf); err != nil {
						// failed to encode
						log.Fatal(err)
					}
					if err := f.Close(); err != nil {
						// failed to close the file
						log.Fatal(err)

					}
					return nil
				},
			},
		},
	}

	nerr := app.Run(os.Args)
	if nerr != nil {
		log.Fatal(nerr)
	}

}
