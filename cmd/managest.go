package cmd

import (
	"context"
	"fmt"
	"strings"
	"strconv"
	"io/ioutil"
	"encoding/json"

	"github.com/spf13/cobra"

	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	roothash "github.com/oasisprotocol/oasis-core/go/roothash/api"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/client"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/connection"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/helpers"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/modules/accounts"
	"github.com/oasisprotocol/oasis-sdk/client-sdk/go/types"

	"github.com/oasisprotocol/cli/cmd/common"
	cliConfig "github.com/oasisprotocol/cli/config"
)

var (

	managestCmd = &cobra.Command{
		Use:   "managest",
		Short: "Manage the stable coin related operations",
	}

	managestShowPropCmd = &cobra.Command{
		Use:   "showproposal [ID]",
		Short: "Show proposal information with proposal ID, latest proposal is output by default",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cliConfig.Global()
			npa := common.GetNPASelection(cfg)				

			if npa.ParaTime != nil {
				// Establish connection with the target network.
				ctx := context.Background()
				c, err := connection.Connect(ctx, npa.Network)
				cobra.CheckErr(err)

				height, err := common.GetActualHeight(
					ctx,
					c.Consensus(),
				)
				cobra.CheckErr(err)

				// Make an effort to support the height query.
				//
				// Note: Public gRPC endpoints do not allow this method.
				round := client.RoundLatest
				if h := common.GetHeight(); h != consensus.HeightLatest {
					blk, err := c.Consensus().RootHash().GetLatestBlock(
						ctx,
						&roothash.RuntimeRequest{
							RuntimeID: npa.ParaTime.Namespace(),
							Height:    height,
						},
					)
					cobra.CheckErr(err)
					round = blk.Header.Round
				}

				var proposalID uint32
				if len(args) > 0 {
					prop_id_input := args[0]
					parsedID, err := strconv.ParseUint(prop_id_input, 10, 32)
					if err != nil {
						// Handle the error, e.g., print an error message and return
						fmt.Printf("Invalid proposal ID: %s\n", prop_id_input)
						return
					}
					// Use the parsed ID
					proposalID = uint32(parsedID)

				}else{
					proposalID, err = c.Runtime(npa.ParaTime).Accounts.ProposalIDInfo(ctx, round)
					cobra.CheckErr(err)
				}


				fmt.Println()
				fmt.Printf("=== %s PARATIME ===\n", npa.ParaTimeName)
				fmt.Printf("Queried proposal ID is: %d. \n", proposalID)

				// Query runtime proposal status when a paratime has been configured.
				proposal, err := c.Runtime(npa.ParaTime).Accounts.ProposalInfo(ctx, round, proposalID)
				cobra.CheckErr(err)

				if proposal.Content.Action != types.NoAction {
					fmt.Printf("===================Proposal============================\n")
					fmt.Printf("Proposal ID: %d\n", proposal.ID)
					fmt.Printf("Proposal Submitter: %s\n", proposal.Submitter.String())
					fmt.Printf("Proposal State: %s\n", proposal.State.String())
					fmt.Printf("Proposal Content:\n")
				    contentStr, err := proposal.Content.String()
					cobra.CheckErr(err)
					for key, value := range contentStr {
						fmt.Printf("    %s: %s\n", key, value)
					}

				    if len(proposal.Results) > 0 {
				        fmt.Println("Results:")
				        for vote, count := range proposal.Results {
				            fmt.Printf("    Vote: %s, Count: %d\n", vote.String(), count)
				        }
				    }
					fmt.Printf("=====================================================\n")
				}



			}else{
				fmt.Println("No paratime specified!")
			}
		},
	}


	managestShowRolesCmd = &cobra.Command{
		Use:   "showroles [role]",
		Short: "Show accounts of specific roles, including Admin, MintProposer, MintVoter etc.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cliConfig.Global()
			npa := common.GetNPASelection(cfg)				

			if npa.ParaTime != nil {
				// Establish connection with the target network.
				ctx := context.Background()
				c, err := connection.Connect(ctx, npa.Network)
				cobra.CheckErr(err)

				height, err := common.GetActualHeight(
					ctx,
					c.Consensus(),
				)
				cobra.CheckErr(err)

				// Make an effort to support the height query.
				//
				// Note: Public gRPC endpoints do not allow this method.
				round := client.RoundLatest
				if h := common.GetHeight(); h != consensus.HeightLatest {
					blk, err := c.Consensus().RootHash().GetLatestBlock(
						ctx,
						&roothash.RuntimeRequest{
							RuntimeID: npa.ParaTime.Namespace(),
							Height:    height,
						},
					)
					cobra.CheckErr(err)
					round = blk.Header.Round
				}



				fmt.Println()
				fmt.Printf("=== %s PARATIME ===\n", npa.ParaTimeName)

				if len(args) > 0 {
					roleStr := args[0]
					role, err := types.RoleFromString(roleStr)
					cobra.CheckErr(err)

					addrs, err := c.Runtime(npa.ParaTime).Accounts.RolesTeam(ctx, round, role)
					if len(addrs) > 0 {
						fmt.Printf("%s: %s\n", roleStr, addrs)
					}

				}else{
					for role := types.Admin; role < types.User; role++ {
						addrs, err := c.Runtime(npa.ParaTime).Accounts.RolesTeam(ctx, round, role)
						cobra.CheckErr(err)
						if len(addrs) > 0 {
							fmt.Printf("%s: %s\n", role.String(), addrs)
						}
					}
				}

			}else{
				fmt.Println("No paratime specified!")
			}
		},
	}


	managestShowQuorumsCmd = &cobra.Command{
		Use:   "showquorums [action]",
		Short: "Show quorums of different actions, including Mint, Burn, SetRoles, Config, etc.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cliConfig.Global()
			npa := common.GetNPASelection(cfg)				

			if npa.ParaTime != nil {
				// Establish connection with the target network.
				ctx := context.Background()
				c, err := connection.Connect(ctx, npa.Network)
				cobra.CheckErr(err)

				height, err := common.GetActualHeight(
					ctx,
					c.Consensus(),
				)
				cobra.CheckErr(err)

				// Make an effort to support the height query.
				//
				// Note: Public gRPC endpoints do not allow this method.
				round := client.RoundLatest
				if h := common.GetHeight(); h != consensus.HeightLatest {
					blk, err := c.Consensus().RootHash().GetLatestBlock(
						ctx,
						&roothash.RuntimeRequest{
							RuntimeID: npa.ParaTime.Namespace(),
							Height:    height,
						},
					)
					cobra.CheckErr(err)
					round = blk.Header.Round
				}



				fmt.Println()
				fmt.Printf("=== %s PARATIME ===\n", npa.ParaTimeName)
				fmt.Printf("Quorums are: \n")

				if len(args) > 0 {
					actionStr := args[0]
					action, err := types.ActionFromString(actionStr)
					cobra.CheckErr(err)

					quorum, err := c.Runtime(npa.ParaTime).Accounts.Quorums(ctx, round, action)
					cobra.CheckErr(err)
					if quorum != 0 {
						fmt.Printf("%s: %d%%\n", actionStr, quorum)
					}

				}else{
					for action := types.SetRoles; action <= types.Config; action++ {
						quorum, err := c.Runtime(npa.ParaTime).Accounts.Quorums(ctx, round, action)
						cobra.CheckErr(err)
						if quorum != 0 {
							fmt.Printf("%s: %d%%\n", action.String(), quorum)
						}
					}
				}

			}else{
				fmt.Println("No paratime specified!")
			}
		},
	}


	managestInitOwnersCmd = &cobra.Command{
		Use:   "initowners [addr1 role1] [addr2 role2] ...",
		Short: "Initialize addresses with roles",
		Long: "Init owners by chain_initiator only one time, roles are [Admin, MintProposer, MintVoter, ...].",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cliConfig.Global()
			npa := common.GetNPASelection(cfg)
			txCfg := common.GetTransactionConfig()

			if npa.Account == nil {
				cobra.CheckErr("no accounts configured in your wallet")
			}

			// When not in offline mode, connect to the given network endpoint.
			ctx := context.Background()
			var conn connection.Connection
			if !txCfg.Offline {
				var err error
				conn, err = connection.Connect(ctx, npa.Network)
				cobra.CheckErr(err)
			}

			var roleAddrs []accounts.RoleAddress

			for i := 0; i < len(args); i += 2 {
				roleAddrStr, roleStr := args[i], args[i+1]

				// Resolve destination address.
				roleAddr, err := common.ResolveLocalAccountOrAddress(npa.Network, roleAddrStr)
				cobra.CheckErr(err)

				role, err := types.RoleFromString(roleStr)
				cobra.CheckErr(err)

				roleAddrs = append(roleAddrs, accounts.RoleAddress{
					Addr: *roleAddr,
					Role: role,
				})
			}

			acc := common.LoadAccount(cfg, npa.AccountName)
			// Prepare transactions for each owner-role pair.

			var sigTx, meta interface{}
			var err error 

			switch npa.ParaTime {
			case nil:
				err := "Invalid paratime configured!"
				cobra.CheckErr(err)

			default:
				// Prepare transaction.
				tx := accounts.NewInitOwnersTx(nil, roleAddrs)

				sigTx, meta, err = common.SignParaTimeTransaction(ctx, npa, acc, conn, tx)
				cobra.CheckErr(err)
			}

			common.BroadcastTransaction(ctx, npa.ParaTime, conn, sigTx, meta, nil)

		},
	}


	managestProposalCmd = &cobra.Command{
		Use:   "propose <proposal.json>",
		Short: "Propose a new proposal with content from a JSON file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cliConfig.Global()
			npa := common.GetNPASelection(cfg)
			txCfg := common.GetTransactionConfig()

			// Read the JSON file.
			jsonFile := args[0]
			jsonData, err := ioutil.ReadFile(jsonFile)
			cobra.CheckErr(err)

		    var raw json.RawMessage
		    tmp := struct {
		        Action string          `json:"action"`
		        Data   *json.RawMessage `json:"data"`
		    }{Data: &raw}

		    err = json.Unmarshal(jsonData, &tmp)
			cobra.CheckErr(err)

			if npa.Account == nil {
				cobra.CheckErr("no accounts configured in your wallet")
			}

			acc := common.LoadAccount(cfg, npa.AccountName)

			// When not in offline mode, connect to the given network endpoint.
			ctx := context.Background()
			var conn connection.Connection
			if !txCfg.Offline {
				var err error
				conn, err = connection.Connect(ctx, npa.Network)
				cobra.CheckErr(err)
			}

			var sigTx, meta interface{}
			switch npa.ParaTime {
			case nil:
				// GB: ignore other layers currently.
			    cobra.CheckErr(fmt.Errorf("Invalid paratime configured!"))

			default:
				action, err := types.ActionFromString(tmp.Action)
				cobra.CheckErr(err)

				// GB: take the input string to dataStr structure.
		        var proposalDataStr *types.ProposalDataStr
		        err = json.Unmarshal(raw, &proposalDataStr)
				cobra.CheckErr(err)
    			// fmt.Printf("gbtest: proposalDataStr is %s\n", proposalDataStr)

				var proposalData types.ProposalData
			    switch action {
			    case types.Mint, types.Burn:
			    	if proposalDataStr.Role != nil || 
			    		proposalDataStr.MintQuorum != nil || proposalDataStr.BurnQuorum != nil || 
			    		proposalDataStr.BlacklistQuorum != nil || proposalDataStr.ConfigQuorum != nil {
			    			cobra.CheckErr(fmt.Errorf("invalid input for proposal!"))
			    	}

			        addr, err := common.ResolveLocalAccountOrAddress(npa.Network, *proposalDataStr.Address)
					cobra.CheckErr(err)
					amtBaseUnits, err := helpers.ParseParaTimeDenomination(npa.ParaTime, *proposalDataStr.Amount, types.NativeDenomination)
					cobra.CheckErr(err)
				    metadata, err := types.StringToMeta(proposalDataStr.Meta)
				    cobra.CheckErr(err)

					proposalData = types.ProposalData{
						Address: addr,
						Amount: amtBaseUnits,
						Meta: metadata,
					}

			    case types.SetRoles:
			    	if proposalDataStr.Amount != nil || 
			    		proposalDataStr.MintQuorum != nil || proposalDataStr.BurnQuorum != nil || 
			    		proposalDataStr.BlacklistQuorum != nil || proposalDataStr.ConfigQuorum != nil {
			    			cobra.CheckErr(fmt.Errorf("invalid input for proposal!"))
			    	}

			        addr, err := common.ResolveLocalAccountOrAddress(npa.Network, *proposalDataStr.Address)
					cobra.CheckErr(err)
			        role, err := types.RoleFromString(*proposalDataStr.Role)
					cobra.CheckErr(err)

					proposalData = types.ProposalData{
						Address: addr,
						Role: &role,
					}
					
			    case types.Whitelist, types.Blacklist:
			    	if proposalDataStr.Role != nil || proposalDataStr.Amount != nil || 
			    		proposalDataStr.MintQuorum != nil || proposalDataStr.BurnQuorum != nil || 
			    		proposalDataStr.BlacklistQuorum != nil || proposalDataStr.ConfigQuorum != nil {
			    			cobra.CheckErr(fmt.Errorf("invalid input for proposal!"))
			    	}

			        addr, err := common.ResolveLocalAccountOrAddress(npa.Network, *proposalDataStr.Address)
					cobra.CheckErr(err)

					proposalData = types.ProposalData{
						Address: addr,
					}

			    case types.Config:
			    	if proposalDataStr.Amount != nil || proposalDataStr.Address != nil {
			    		cobra.CheckErr(fmt.Errorf("invalid input for proposal!"))
			    	}

					proposalData = types.ProposalData{
						MintQuorum: proposalDataStr.MintQuorum,
					    BurnQuorum: proposalDataStr.BurnQuorum,
						WhitelistQuorum: proposalDataStr.WhitelistQuorum, 
					    BlacklistQuorum: proposalDataStr.BlacklistQuorum,
					    ConfigQuorum: proposalDataStr.ConfigQuorum,
					}

			    default:
			    	cobra.CheckErr(fmt.Errorf("invalid action for proposal!"))
			    }

				// Prepare transaction.
				proposal := &accounts.ProposalContent{
					Action: action,
					Data: proposalData,
				}				
				tx := accounts.NewProposeTx(nil, proposal)

				sigTx, meta, err = common.SignParaTimeTransaction(ctx, npa, acc, conn, tx)
				cobra.CheckErr(err)
			}

			common.BroadcastTransaction(ctx, npa.ParaTime, conn, sigTx, meta, nil)
		},
	}

	// GB: insert MintST (stable coin) function command 
	managestVoteCmd = &cobra.Command{
		Use:   "vote <proposalID> <option>",
		Short: "Vote the proposal with options of YES, NO and ABSTAIN.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cliConfig.Global()
			npa := common.GetNPASelection(cfg)
			txCfg := common.GetTransactionConfig()
			proposalID, option := args[0], args[1]

			if npa.Account == nil {
				cobra.CheckErr("no accounts configured in your wallet")
			}

			// When not in offline mode, connect to the given network endpoint.
			ctx := context.Background()
			var conn connection.Connection
			if !txCfg.Offline {
				var err error
				conn, err = connection.Connect(ctx, npa.Network)
				cobra.CheckErr(err)
			}

			acc := common.LoadAccount(cfg, npa.AccountName)

			var sigTx, meta interface{}
			switch npa.ParaTime {
			case nil:
			    cobra.CheckErr(fmt.Errorf("Invalid paratime configured!"))

			default:
				u64ID, err := strconv.ParseUint(proposalID, 10, 32)
				u32ID := uint32(u64ID)

    			lowercaseOption := strings.ToLower(option)
    			voteOp, err := types.StringToVote(lowercaseOption)
				cobra.CheckErr(err)

				// Prepare transaction.
				tx := accounts.NewVoteSTTx(nil, &accounts.VoteProposal{
					ID:     u32ID,
					Option: voteOp,
				})

				sigTx, meta, err = common.SignParaTimeTransaction(ctx, npa, acc, conn, tx)
				cobra.CheckErr(err)
			}

			common.BroadcastTransaction(ctx, npa.ParaTime, conn, sigTx, meta, nil)
		},
	}


)


func init() {
	managestShowPropCmd.Flags().AddFlagSet(common.SelectorFlags)
	managestShowPropCmd.Flags().AddFlagSet(common.HeightFlag)
	managestShowRolesCmd.Flags().AddFlagSet(common.SelectorFlags)
	managestShowRolesCmd.Flags().AddFlagSet(common.HeightFlag)
	managestShowQuorumsCmd.Flags().AddFlagSet(common.SelectorFlags)
	managestShowQuorumsCmd.Flags().AddFlagSet(common.HeightFlag)


	managestInitOwnersCmd.Flags().AddFlagSet(common.SelectorFlags)
	managestInitOwnersCmd.Flags().AddFlagSet(common.TransactionFlags)
	managestInitOwnersCmd.Flags().AddFlagSet(common.ForceFlag)

	managestProposalCmd.Flags().AddFlagSet(common.SelectorFlags)
	managestProposalCmd.Flags().AddFlagSet(common.TransactionFlags)
	managestProposalCmd.Flags().AddFlagSet(common.ForceFlag)


	managestVoteCmd.Flags().AddFlagSet(common.SelectorFlags)
	managestVoteCmd.Flags().AddFlagSet(common.TransactionFlags)
	managestVoteCmd.Flags().AddFlagSet(common.ForceFlag)



	managestCmd.AddCommand(managestShowPropCmd)	
	managestCmd.AddCommand(managestShowRolesCmd)	
	managestCmd.AddCommand(managestShowQuorumsCmd)	

	managestCmd.AddCommand(managestInitOwnersCmd)	
	managestCmd.AddCommand(managestProposalCmd)	
	managestCmd.AddCommand(managestVoteCmd)	

}





