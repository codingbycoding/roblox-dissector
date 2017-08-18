package main
import "github.com/google/gopacket"
import "github.com/davecgh/go-spew/spew"
import "errors"
import "strconv"

var Vector3Override = map[string]struct{}{
	"Rotation": struct{}{},
	"CenterOfMass": struct{}{},
	"OrientationLocal": struct{}{},
	"Orientation": struct{}{},
	"PositionLocal": struct{}{},
	"Position": struct{}{},
	"RotVelocity": struct{}{},
	"size": struct{}{},
	"Size": struct{}{},
	"Velocity": struct{}{},
	"siz": struct{}{}, // ???
}

type ReplicationProperty struct {
	Name string
	Type string
	Value PropertyValue
	IsDefault bool
}

type PropertyValue interface {
	Show() string
}

func (this *ReplicationProperty) Show() string {
	if this == nil {
		return "ERR!! NIL"
	}
	if this.IsDefault {
		return "!DEFAULT"
	}
	return this.Value.Show()
}

const (
	ROUND_JOINDATA	= iota
	ROUND_STRINGS	= iota
	ROUND_OTHER		= iota
	ROUND_UPDATE	= iota
)

func (schema *PropertySchemaItem) Decode(round int, thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (*ReplicationProperty, error) {
	var err error
	if !schema.Replicates && round != ROUND_UPDATE {
		return nil, nil
	}
	isJoinData := round == ROUND_JOINDATA

	isStringObject := true
	if schema.Type != "Object" &&
	schema.Type != "string" &&
	schema.Type != "ProtectedString" &&
	schema.Type != "BinaryString" &&
	schema.Type != "SystemAddress" &&
	schema.Type != "Content" {
		isStringObject = false
	}

	if round == ROUND_STRINGS && !isStringObject {
		return nil, nil
	}
	if round == ROUND_OTHER && isStringObject {
		return nil, nil
	}

	Property := &ReplicationProperty{Type: schema.Type, Name: schema.Name}
	if schema.Type == "bool" {
		if err != nil {
			return Property, err
		}
		Property.Value, err = thisBitstream.ReadPBool()
		println(DebugInfo2(context, packet, isJoinData), "Read bool", schema.Name, bool(Property.Value.(pbool)))
		if err != nil {
			return Property, err
		}
	} else {
		if round != ROUND_UPDATE {
			Property.IsDefault, err = thisBitstream.ReadBool()
			if err != nil {
				return Property, err
			}
			if Property.IsDefault {
				println(DebugInfo2(context, packet, isJoinData), "Read", schema.Name, "1 bit: default")
				return Property, nil
			}
		}
		switch schema.Type {
		case "string":
			Property.Value, err = thisBitstream.ReadPString(isJoinData, context)
			break
		case "ProtectedString":
			Property.Value, err = thisBitstream.ReadProtectedString(isJoinData, context)
			break
		case "BinaryString":
			Property.Value, err = thisBitstream.ReadBinaryString()
			break
		case "int":
			Property.Value, err = thisBitstream.ReadPSInt()
			break
		case "float":
			Property.Value, err = thisBitstream.ReadPFloat()
			break
		case "double":
			Property.Value, err = thisBitstream.ReadPDouble()
			break
		case "Axes":
			Property.Value, err = thisBitstream.ReadAxes()
			break
		case "Faces":
			Property.Value, err = thisBitstream.ReadFaces()
			break
		case "BrickColor":
			Property.Value, err = thisBitstream.ReadBrickColor()
			break
		case "Object":
			Property.Value, err = thisBitstream.ReadObject(isJoinData, context)
			break
		case "UDim":
			Property.Value, err = thisBitstream.ReadUDim()
			break
		case "UDim2":
			Property.Value, err = thisBitstream.ReadUDim2()
			break
		case "Vector2":
			Property.Value, err = thisBitstream.ReadVector2()
			break
		case "Vector3":
			if _, ok := Vector3Override[schema.Name]; ok {
				Property.Value, err = thisBitstream.ReadVector3()
			} else {
				Property.Value, err = thisBitstream.ReadVector3Simple()
			}

			break
		case "Vector2uint16":
			Property.Value, err = thisBitstream.ReadVector2uint16()
			break
		case "Vector3uint16":
			Property.Value, err = thisBitstream.ReadVector3uint16()
			break
		case "Ray":
			Property.Value, err = thisBitstream.ReadRay()
			break
		case "Color3":
			Property.Value, err = thisBitstream.ReadColor3()
			break
		case "Color3uint8":
			Property.Value, err = thisBitstream.ReadColor3uint8()
			break
		case "CoordinateFrame":
			Property.Value, err = thisBitstream.ReadCFrame()
			break
		case "Content":
			Property.Value, err = thisBitstream.ReadContent(isJoinData, context)
			break
		case "SystemAddress":
			Property.Value, err = thisBitstream.ReadSystemAddress(isJoinData, context)
			break
		default:
			if schema.IsEnum {
				Property.Value, err = thisBitstream.ReadEnumValue(schema.BitSize)
			} else {
				return Property, errors.New("property parser encountered unknown type: " + schema.Type)
			}
		}
		if schema.Type != "ProtectedString" {
			println(DebugInfo2(context, packet, isJoinData), "Read", schema.Name, spew.Sdump(Property.Value))
		} else {
			println(DebugInfo2(context, packet, isJoinData), "Read", schema.Name, len(Property.Value.(ProtectedString)), spew.Sdump(Property.Value.(ProtectedString)[:0x0a]))
		}
	}
	return Property, nil
}

func (schema StaticPropertySchema) Decode(round int, thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, isRebind bool) (*ReplicationProperty, error) {
	var err error
	result := &ReplicationProperty{schema.Name, schema.TypeString, nil, false}
	isJoinData := round == ROUND_JOINDATA
	if round != ROUND_UPDATE {
		result.IsDefault, err = thisBitstream.ReadBool()
		if result.IsDefault || err != nil {
			println(DebugInfo2(context, packet, isJoinData), "Read", schema.Name, "default")
			return result, err
		}
	}

	switch schema.Type {
	case PROP_TYPE_STRING:
		result.Value, err = thisBitstream.ReadNewPString(isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		result.Value, err = thisBitstream.ReadNewPString(true, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		result.Value, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		result.Value, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		result.Value, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		result.Value, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_ENUM:
		result.Value, err = thisBitstream.ReadNewEnumValue()
	case PROP_TYPE_BINARYSTRING:
		result.Value, err = thisBitstream.ReadNewBinaryString()
	case PROP_TYPE_PBOOL:
		result.Value, err = thisBitstream.ReadPBool()
	case PROP_TYPE_PSINT:
		result.Value, err = thisBitstream.ReadNewPSint()
	case PROP_TYPE_PFLOAT:
		result.Value, err = thisBitstream.ReadPFloat()
	case PROP_TYPE_PDOUBLE:
		result.Value, err = thisBitstream.ReadPDouble()
	case PROP_TYPE_UDIM:
		result.Value, err = thisBitstream.ReadUDim()
	case PROP_TYPE_UDIM2:
		result.Value, err = thisBitstream.ReadUDim2()
	case PROP_TYPE_RAY:
		result.Value, err = thisBitstream.ReadRay()
	case PROP_TYPE_FACES:
		result.Value, err = thisBitstream.ReadFaces()
	case PROP_TYPE_AXES:
		result.Value, err = thisBitstream.ReadAxes()
	case PROP_TYPE_BRICKCOLOR:
		result.Value, err = thisBitstream.ReadBrickColor()
	case PROP_TYPE_COLOR3:
		result.Value, err = thisBitstream.ReadColor3()
	case PROP_TYPE_COLOR3UINT8:
		result.Value, err = thisBitstream.ReadColor3uint8()
	case PROP_TYPE_VECTOR2:
		result.Value, err = thisBitstream.ReadVector2()
	case PROP_TYPE_VECTOR3_SIMPLE:
		result.Value, err = thisBitstream.ReadVector3Simple()
	case PROP_TYPE_VECTOR3_COMPLICATED:
		result.Value, err = thisBitstream.ReadVector3()
	case PROP_TYPE_VECTOR2UINT16:
		result.Value, err = thisBitstream.ReadVector2uint16()
	case PROP_TYPE_VECTOR3UINT16:
		result.Value, err = thisBitstream.ReadVector3uint16()
	case PROP_TYPE_CFRAME_SIMPLE:
		result.Value, err = thisBitstream.ReadCFrameSimple()
	case PROP_TYPE_CFRAME_COMPLICATED:
		result.Value, err = thisBitstream.ReadCFrame()
	case PROP_TYPE_INSTANCE:
		result.Value, err = thisBitstream.ReadObject(isJoinData, context)
	case PROP_TYPE_CONTENT:
		result.Value, err = thisBitstream.ReadNewContent(isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		result.Value, err = thisBitstream.ReadSystemAddress(isJoinData, context)
	default:
		return nil, errors.New("Unsupported property type: " + schema.TypeString + strconv.Itoa(int(schema.Type)))
	}
	if schema.TypeString != "ProtectedString" {
		println(DebugInfo2(context, packet, isJoinData), "Read", schema.Name, spew.Sdump(result.Value))
	} else {
		println(DebugInfo2(context, packet, isJoinData), "Read", schema.Name, len(result.Value.(ProtectedString)))
	}
	return result, err
}
