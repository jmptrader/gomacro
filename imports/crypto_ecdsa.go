// this file was generated by gomacro command: import "crypto/ecdsa"
// DO NOT EDIT! Any change will be lost when the file is re-generated

package imports

import (
	pkg "crypto/ecdsa"
	. "reflect"
)

func Package_crypto_ecdsa() (map[string]Value, map[string]Type) {
	return map[string]Value{
			"GenerateKey": ValueOf(pkg.GenerateKey),
			"Sign":        ValueOf(pkg.Sign),
			"Verify":      ValueOf(pkg.Verify),
		}, map[string]Type{
			"PrivateKey": TypeOf((*pkg.PrivateKey)(nil)).Elem(),
			"PublicKey":  TypeOf((*pkg.PublicKey)(nil)).Elem(),
		}
}

func init() {
	binds, types := Package_crypto_ecdsa()
	Binds["crypto/ecdsa"] = binds
	Types["crypto/ecdsa"] = types
}