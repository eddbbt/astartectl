// Copyright © 2019 Ispirata Srl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package realm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// triggersCmd represents the triggers command
var triggersCmd = &cobra.Command{
	Use:     "triggers",
	Short:   "Manage triggers",
	Long:    `List, show, install or delete triggers in your realm.`,
	Aliases: []string{"trigger"},
}

var triggersListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List triggers",
	Long:    `List the name of triggers installed in the realm.`,
	Example: `  astartectl realm-management triggers list`,
	RunE:    triggersListF,
	Aliases: []string{"ls"},
}

var triggersShowCmd = &cobra.Command{
	Use:     "show <trigger_name>",
	Short:   "Show trigger",
	Long:    `Shows a trigger installed in the realm.`,
	Example: `  astartectl realm-management triggers show my_data_trigger`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersShowF,
}

var triggersInstallCmd = &cobra.Command{
	Use:   "install <trigger_file>",
	Short: "Install trigger",
	Long: `Install the given trigger in the realm.
<trigger_file> must be a path to a JSON file containing a valid Astarte trigger.`,
	Example: `  astartectl realm-management triggers install my_data_trigger.json`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersInstallF,
}

var triggersDeleteCmd = &cobra.Command{
	Use:     "delete <trigger_name>",
	Short:   "Delete a trigger",
	Long:    `Deletes the specified trigger from the realm.`,
	Example: `  astartectl realm-management triggers delete my_data_trigger`,
	Args:    cobra.ExactArgs(1),
	RunE:    triggersDeleteF,
	Aliases: []string{"del"},
}

var triggersSaveCmd = &cobra.Command{
	Use:   "save [destination-path]",
	Short: "Save triggers to a local folder",
	Long: `Save each trigger in a realm to a local folder. Each trigger will
be saved in a dedicated file whose name will be in the form '<trigger_name>_v<version>.json'.
When no destination path is set, triggers will be saved in the current working directory.
This command does not support the --to-curl flag.`,
	Example: `  astartectl realm-management triggers save`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    triggersSaveF,
}

func init() {
	RealmManagementCmd.AddCommand(triggersCmd)

	triggersCmd.AddCommand(
		triggersListCmd,
		triggersShowCmd,
		triggersInstallCmd,
		triggersDeleteCmd,
		triggersSaveCmd,
	)
}

func triggersListF(command *cobra.Command, args []string) error {
	realmTriggersCall, err := astarteAPIClient.ListTriggers(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(realmTriggersCall, astarteAPIClient)

	realmTriggersRes, err := realmTriggersCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rawRealmTriggers, _ := realmTriggersRes.Parse()
	realmTriggers, _ := rawRealmTriggers.([]string)
	fmt.Println(realmTriggers)
	return nil
}

func triggersShowF(command *cobra.Command, args []string) error {
	triggerName := args[0]

	getTriggerCall, err := astarteAPIClient.GetTrigger(realm, triggerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(getTriggerCall, astarteAPIClient)

	getTriggerRes, err := getTriggerCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	triggerDefinition, _ := getTriggerRes.Parse()
	respJSON, _ := json.MarshalIndent(triggerDefinition, "", "  ")
	fmt.Println(string(respJSON))

	return nil
}

func triggersInstallF(command *cobra.Command, args []string) error {
	triggerFile, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	var triggerBody map[string]interface{}
	err = json.Unmarshal(triggerFile, &triggerBody)
	if err != nil {
		return err
	}

	installTriggerCall, err := astarteAPIClient.InstallTrigger(realm, triggerBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(installTriggerCall, astarteAPIClient)

	installTriggerRes, err := installTriggerCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = installTriggerRes.Parse()

	fmt.Println("ok")
	return nil
}

func triggersDeleteF(command *cobra.Command, args []string) error {
	triggerName := args[0]
	deleteTriggerCall, err := astarteAPIClient.DeleteTrigger(realm, triggerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	utils.MaybeCurlAndExit(deleteTriggerCall, astarteAPIClient)

	deleteTriggerRes, err := deleteTriggerCall.Run(astarteAPIClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = deleteTriggerRes.Parse()

	fmt.Println("ok")
	return nil
}

func triggersSaveF(command *cobra.Command, args []string) error {
	if viper.GetBool("realmmanagement-to-curl") {
		fmt.Println(`'triggers save' does not support the --to-curl option. Use 'triggers list' to get the triggers in your realm, 'triggers versions' to get their versions, and 'triggers show' to get the content of an interface.`)
		os.Exit(1)
	}

	var targetPath string
	var err error
	if len(args) == 0 {
		targetPath, _ = filepath.Abs(".")
	} else {
		targetPath, err = filepath.Abs(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// retrieve triggers list
	realmTriggers, err := ListTriggers(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	//ifaceNameAndVersions := map[string][]int{}

	// // and the versions for each interface
	// for _, ifaceName := range realmTriggers {
	// 	interfaceVersions, err := interfaceVersions(ifaceName)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 		os.Exit(1)
	// 	}
	// 	ifaceNameAndVersions[ifaceName] = ifaceName
	// }

	for _, name := range realmTriggers {

		triggerDefinition, err := getTriggerDefinition(realm, name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		respJSON, err := json.MarshalIndent(triggerDefinition, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("/%s/%s.json", targetPath, name)
		outFile, err := os.Create(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer outFile.Close()

		if _, err := outFile.Write(respJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	}
	return nil
}
func ListTriggers(realm string) ([]string, error) {
	listTriggersCall, err := astarteAPIClient.ListTriggers(realm)
	if err != nil {
		return []string{}, err
	}

	utils.MaybeCurlAndExit(listTriggersCall, astarteAPIClient)

	listTriggersRes, err := listTriggersCall.Run(astarteAPIClient)
	if err != nil {
		return []string{}, err
	}
	rawlistTriggers, err := listTriggersRes.Parse()
	if err != nil {
		return []string{}, err
	}
	return rawlistTriggers.([]string), nil
}
func getTriggerDefinition(realm, triggerName string) (map[string]interface{}, error) {
	getTriggerCall, err := astarteAPIClient.GetTrigger(realm, triggerName)
	if err != nil {
		return nil, err
	}

	// When we're here in the context of `interfaces sync`, the to-curl flag
	// is always false (`interfaces sync` has no `--to-curl` flag)
	// and thus the call will never exit unexpectedly
	utils.MaybeCurlAndExit(getTriggerCall, astarteAPIClient)

	getTriggerRes, err := getTriggerCall.Run(astarteAPIClient)
	if err != nil {
		return nil, err
	}
	rawTRigger, err := getTriggerRes.Parse()
	if err != nil {
		return nil, err
	}
	triggerDefinition, _ := rawTRigger.(map[string]interface{})
	return triggerDefinition, nil
}
