package keys

import (
    "bytes"
    "strings"
    "testing"

    "github.com/stretchr/testify/require"
    sdk "github.com/baron-chain/cosmos-sdk/types"
)

func TestParseKey(t *testing.T) {
    testCases := []struct {
        name        string
        input       string
        format      string
        expectError bool
        check       func(*testing.T, string)
    }{
        {
            name:        "empty input",
            input:       "",
            expectError: true,
        },
        {
            name:        "invalid input",
            input:       "invalid",
            expectError: true,
        },
        {
            name:  "classic bech32 address",
            input: "baron104ytdpvrx9284zd50v9ep8c6j7pua7dkk0x3ek",
            check: func(t *testing.T, output string) {
                require.Contains(t, output, "Human readable part: baron")
                require.Contains(t, output, "Bytes (hex):")
            },
        },
        {
            name:  "quantum-safe kyber address",
            input: "baronkyber1qw8ausxvhes3l9v0y5vh0lqz3tyqk4vr2x9a",
            check: func(t *testing.T, output string) {
                require.Contains(t, output, "Human readable part: baronkyber")
                require.Contains(t, output, "Bytes (hex):")
            },
        },
        {
            name:  "quantum-safe dilithium address",
            input: "barondilithium1hj4ry9h5z5vx0y5vh0lqz3tyqk4vr2x9a",
            check: func(t *testing.T, output string) {
                require.Contains(t, output, "Human readable part: barondilithium")
                require.Contains(t, output, "Bytes (hex):")
            },
        },
        {
            name:  "hex address",
            input: "EB5AE9872103497EC092EF901027049E4F39200C60040D3562CD7F104A39F62E6E5A39A818F4",
            check: func(t *testing.T, output string) {
                require.Contains(t, output, "baron1")
                require.Contains(t, output, "baronvaloper1")
            },
        },
        {
            name:   "json output format",
            input:  "baron104ytdpvrx9284zd50v9ep8c6j7pua7dkk0x3ek",
            format: "json",
            check: func(t *testing.T, output string) {
                require.Contains(t, output, `"human_readable": "baron"`)
                require.Contains(t, output, `"hex_bytes":`)
            },
        },
        {
            name:   "yaml output format",
            input:  "baron104ytdpvrx9284zd50v9ep8c6j7pua7dkk0x3ek",
            format: "yaml",
            check: func(t *testing.T, output string) {
                require.Contains(t, output, "human_readable: baron")
                require.Contains(t, output, "hex_bytes:")
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := ParseKeyStringCommand()
            var buf bytes.Buffer
            cmd.SetOut(&buf)

            if tc.format != "" {
                cmd.SetArgs([]string{tc.input, "--format", tc.format})
            } else {
                cmd.SetArgs([]string{tc.input})
            }

            err := cmd.Execute()

            if tc.expectError {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            if tc.check != nil {
                tc.check(t, buf.String())
            }
        })
    }
}

func TestBech32Prefixes(t *testing.T) {
    config := sdk.NewConfig()
    config.SetBech32PrefixForAccount("baron", "baronpub")
    config.SetBech32PrefixForValidator("baronvaloper", "baronvaloperpub")
    config.SetBech32PrefixForConsensusNode("baronvalcons", "baronvalconspub")

    prefixes := getBech32Prefixes(config)
    require.Equal(t, 6, len(prefixes))
    require.Equal(t, "baron", prefixes[0])
    require.Equal(t, "baronpub", prefixes[1])
    require.Equal(t, "baronvaloper", prefixes[2])
}

func TestOutputFormats(t *testing.T) {
    testCases := []struct {
        name     string
        format   string
        input    string
        contains []string
    }{
        {
            name:   "text format",
            format: "text",
            input:  "baron104ytdpvrx9284zd50v9ep8c6j7pua7dkk0x3ek",
            contains: []string{
                "Human readable part:",
                "Bytes (hex):",
            },
        },
        {
            name:   "json format",
            format: "json",
            input:  "baron104ytdpvrx9284zd50v9ep8c6j7pua7dkk0x3ek",
            contains: []string{
                `"human_readable":`,
                `"hex_bytes":`,
            },
        },
        {
            name:   "yaml format",
            format: "yaml",
            input:  "baron104ytdpvrx9284zd50v9ep8c6j7pua7dkk0x3ek",
            contains: []string{
                "human_readable:",
                "hex_bytes:",
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := ParseKeyStringCommand()
            var buf bytes.Buffer
            cmd.SetOut(&buf)
            cmd.SetArgs([]string{tc.input, "--format", tc.format})

            err := cmd.Execute()
            require.NoError(t, err)

            output := buf.String()
            for _, s := range tc.contains {
                require.True(t, strings.Contains(output, s))
            }
        })
    }
}
