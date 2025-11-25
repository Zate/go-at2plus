package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zberg/go-at2plus/pkg/at2plus"
)

var (
	targetIP string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&targetIP, "ip", "", "IP address of the AirTouch 2+ unit")

	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(controlGroupCmd)
	rootCmd.AddCommand(controlACCmd)
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover AirTouch 2+ devices on the network",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Discovering devices...")
		results, err := at2plus.Discover()
		if err != nil {
			fmt.Printf("Error discovering: %v\n", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("No devices found.")
			return
		}

		for _, res := range results {
			fmt.Printf("Found device at: %s\n", res.IP)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of groups and ACs",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		defer client.Close()

		fmt.Println("Fetching Group Status...")
		groups, err := client.GetGroupStatus()
		if err != nil {
			fmt.Printf("Error getting group status: %v\n", err)
		} else {
			for _, g := range groups {
				powerStr := "OFF"
				if g.Power == 1 {
					powerStr = "ON"
				}
				if g.Power == 3 {
					powerStr = "TURBO"
				}

				fmt.Printf("Group %d: Power=%s, Open=%d%%\n", g.GroupNumber, powerStr, g.Percent)
			}
		}

		fmt.Println("\nFetching AC Status...")
		acs, err := client.GetACStatus()
		if err != nil {
			fmt.Printf("Error getting AC status: %v\n", err)
		} else {
			for _, ac := range acs {
				powerStr := "OFF"
				if ac.Power == 1 {
					powerStr = "ON"
				}

				modeStr := "AUTO"
				switch ac.Mode {
				case 1:
					modeStr = "HEAT"
				case 2:
					modeStr = "DRY"
				case 3:
					modeStr = "FAN"
				case 4:
					modeStr = "COOL"
				}

				fmt.Printf("AC %d: Power=%s, Mode=%s, Temp=%d, Setpoint=%d\n", ac.ACNumber, powerStr, modeStr, ac.Temperature, ac.Setpoint)
			}
		}
	},
}

var controlGroupCmd = &cobra.Command{
	Use:   "control-group [group-number]",
	Short: "Control a group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupNum, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Invalid group number '%s': must be a number\n", args[0])
			os.Exit(1)
		}
		if groupNum < 0 || groupNum > 15 {
			fmt.Printf("Invalid group number %d: must be 0-15\n", groupNum)
			os.Exit(1)
		}

		powerStr, _ := cmd.Flags().GetString("power")
		percent, _ := cmd.Flags().GetInt("percent")

		var power *int
		if powerStr != "" {
			p := 0 // Next
			if powerStr == "off" {
				p = 1
			}
			if powerStr == "on" {
				p = 2
			}
			if powerStr == "turbo" {
				p = 3
			}
			power = &p
		}

		var pct *int
		if cmd.Flags().Changed("percent") {
			pct = &percent
		}

		client := getClient()
		defer client.Close()

		err = client.SetGroupControl([]at2plus.GroupControl{
			{
				GroupNumber: uint8(groupNum),
				Power:       power,
				Percent:     pct,
			},
		})

		if err != nil {
			fmt.Printf("Error controlling group: %v\n", err)
		} else {
			fmt.Println("Command sent successfully.")
		}
	},
}

var controlACCmd = &cobra.Command{
	Use:   "control-ac [ac-number]",
	Short: "Control an AC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		acNum, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Invalid AC number '%s': must be a number\n", args[0])
			os.Exit(1)
		}
		if acNum < 0 || acNum > 7 {
			fmt.Printf("Invalid AC number %d: must be 0-7\n", acNum)
			os.Exit(1)
		}

		powerStr, _ := cmd.Flags().GetString("power")
		modeStr, _ := cmd.Flags().GetString("mode")
		temp, _ := cmd.Flags().GetInt("temp")

		var power *int
		if powerStr != "" {
			p := 1 // Toggle
			if powerStr == "off" {
				p = 2
			}
			if powerStr == "on" {
				p = 3
			}
			power = &p
		}

		var mode *int
		if modeStr != "" {
			m := 0 // Auto
			if modeStr == "heat" {
				m = 1
			}
			if modeStr == "dry" {
				m = 2
			}
			if modeStr == "fan" {
				m = 3
			}
			if modeStr == "cool" {
				m = 4
			}
			mode = &m
		}

		var setpoint *int
		if cmd.Flags().Changed("temp") {
			setpoint = &temp
		}

		client := getClient()
		defer client.Close()

		err = client.SetACControl([]at2plus.ACControl{
			{
				ACNumber: uint8(acNum),
				Power:    power,
				Mode:     mode,
				Setpoint: setpoint,
			},
		})

		if err != nil {
			fmt.Printf("Error controlling AC: %v\n", err)
		} else {
			fmt.Println("Command sent successfully.")
		}
	},
}

func init() {
	controlGroupCmd.Flags().String("power", "", "Power state (on, off, turbo)")
	controlGroupCmd.Flags().Int("percent", 0, "Open percentage (0-100)")

	controlACCmd.Flags().String("power", "", "Power state (on, off)")
	controlACCmd.Flags().String("mode", "", "Mode (auto, heat, dry, fan, cool)")
	controlACCmd.Flags().Int("temp", 0, "Temperature setpoint")
}

func getClient() *at2plus.Client {
	if targetIP == "" {
		fmt.Println("IP address required. Use --ip flag or run discover first.")
		os.Exit(1)
	}

	client, err := at2plus.NewClient(targetIP)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v\n", targetIP, err)
		os.Exit(1)
	}
	return client
}
