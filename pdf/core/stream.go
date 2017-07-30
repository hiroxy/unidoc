/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"errors"
	"fmt"

	"github.com/unidoc/unidoc/common"
)

// NewEncoderFromStream creates an encoder from `streamObj`'s dictionary.
func NewEncoderFromStream(streamObj *PdfObjectStream) (StreamEncoder, error) {
	filterObj := streamObj.PdfObjectDictionary.Get("Filter")
	if filterObj == nil {
		// No filter, return raw data back.
		return NewRawEncoder(), nil
	}

	if _, isNull := filterObj.(*PdfObjectNull); isNull {
		// Filter is null -> raw data.
		return NewRawEncoder(), nil
	}

	// The filter should be a name or an array with a list of filter names.
	method, ok := filterObj.(*PdfObjectName)
	if !ok {
		array, ok := filterObj.(*PdfObjectArray)
		if !ok {
			return nil, errors.New("Filter not a Name or Array object")
		}
		if len(*array) == 0 {
			// Empty array -> indicates raw filter (no filter).
			return NewRawEncoder(), nil
		}

		if len(*array) != 1 {
			menc, err := newMultiEncoderFromStream(streamObj)
			if err != nil {
				common.Log.Error("Failed creating multi encoder: %v", err)
				return nil, err
			}

			common.Log.Trace("Multi enc: %s", menc)
			return menc, nil
		}

		// Single element.
		filterObj = (*array)[0]
		method, ok = filterObj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Filter array member not a Name object")
		}
	}
	// !@#$ switch
	if *method == StreamEncodingFilterNameFlate {
		return newFlateEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameLZW {
		return newLZWEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameDCT {
		return newDCTEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameRunLength {
		return newRunLengthEncoderFromStream(streamObj, nil)
	} else if *method == StreamEncodingFilterNameASCIIHex {
		return NewASCIIHexEncoder(), nil
	} else if *method == StreamEncodingFilterNameASCII85 {
		return NewASCII85Encoder(), nil
	} else if *method == StreamEncodingFilterNameCCITTFax {
		return NewCCITTFaxEncoder(), nil
	} else if *method == StreamEncodingFilterNameJBIG2 {
		return NewJBIG2Encoder(), nil
	} else if *method == StreamEncodingFilterNameJPX {
		return NewJPXEncoder(), nil
	} else {
		err := fmt.Errorf("Unsupported encoding method (%s)", *method)
		common.Log.Error("err=%v", err)
		panic(err)
		return nil, err
	}
}

// Decodes the stream.
// Supports FlateDecode, ASCIIHexDecode, LZW.
func DecodeStream(streamObj *PdfObjectStream) ([]byte, error) {
	common.Log.Trace("Decode stream")

	encoder, err := NewEncoderFromStream(streamObj)
	if err != nil {
		common.Log.Error("Stream decoding failed: %v", err)
		return nil, err
	}
	common.Log.Trace("Encoder: %#v", encoder)

	decoded, err := encoder.DecodeStream(streamObj)
	if err != nil {
		common.Log.Error("Stream decoding failed: %v", err)
		// panic(err)
		return nil, err
	}

	return decoded, nil
}

// Encodes the stream.
// Uses the encoding specified by the object.
func EncodeStream(streamObj *PdfObjectStream) error {
	common.Log.Trace("Encode stream")

	encoder, err := NewEncoderFromStream(streamObj)
	if err != nil {
		common.Log.Debug("Stream decoding failed: %v", err)
		return err
	}

	if lzwenc, is := encoder.(*LZWEncoder); is {
		// If LZW:
		// Make sure to use EarlyChange 0.. We do not have write support for 1 yet.
		lzwenc.EarlyChange = 0
		streamObj.PdfObjectDictionary.Set("EarlyChange", MakeInteger(0))
	}

	common.Log.Trace("Encoder: %+v\n", encoder)
	encoded, err := encoder.EncodeBytes(streamObj.Stream)
	if err != nil {
		common.Log.Debug("Stream encoding failed: %v", err)
		return err
	}

	streamObj.Stream = encoded

	// Update length
	streamObj.PdfObjectDictionary.Set("Length", MakeInteger(int64(len(encoded))))

	return nil
}
