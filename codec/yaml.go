package codec

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"sigs.k8s.io/yaml"
)

// ErrNilMessage is returned when attempting to marshal a nil proto.Message
var ErrNilMessage = fmt.Errorf("cannot marshal nil proto.Message")

// ErrNilCodec is returned when attempting to marshal with a nil JSONCodec
var ErrNilCodec = fmt.Errorf("cannot marshal with nil JSONCodec")

// MarshalYAML marshals a protobuf message to YAML format using the provided JSONCodec.
// It first marshals the message to JSON using the codec's specialized MarshalJSON methods,
// then converts the JSON to YAML. This approach ensures proper handling of protobuf and amino
// serialization based on configuration, though it requires an additional JSON conversion step.
//
// Parameters:
//   - cdc: JSONCodec for initial JSON marshaling
//   - msg: proto.Message to be marshaled
//
// Returns:
//   - []byte: YAML-encoded data
//   - error: any error encountered during marshaling
//
// Note: This function is not optimized for performance and should not be used in
// performance-critical paths.
func MarshalYAML(cdc JSONCodec, msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, ErrNilMessage
	}

	if cdc == nil {
		return nil, ErrNilCodec
	}

	// Marshal to JSON first using the codec's specialized methods
	jsonData, err := cdc.MarshalJSON(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	// Convert JSON to YAML
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}

	return yamlData, nil
}
