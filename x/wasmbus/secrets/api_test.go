package secrets

import (
	"encoding/json"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/nats-io/nkeys"
)

func TestJWTClaims(t *testing.T) {
	claims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "test",
		Subject:   "somebody",
		ID:        "1",
		Audience:  []string{"somebody_else"},
	}

	kp, err := nkeys.CreateAccount()
	if err != nil {
		t.Fatal(err)
	}

	token := jwt.NewWithClaims(SigningMethodEd25519, claims)
	_, err = token.SignedString(kp)
	if err != nil {
		t.Fatal(err)
	}

	validJWT := "eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJTakI1Zm05NzRTanU5V01nVFVjaHNiIiwiaWF0IjoxNjQ0ODQzNzQzLCJpc3MiOiJBQ09KSk42V1VQNE9ERDc1WEVCS0tUQ0NVSkpDWTVaS1E1NlhWS1lLNEJFSldHVkFPT1FIWk1DVyIsInN1YiI6Ik1CQ0ZPUE02SlcyQVBKTFhKRDNaNU80Q043Q1BZSjJCNEZUS0xKVVI1WVI1TUlUSVU3SEQzV0Q1Iiwid2FzY2FwIjp7Im5hbWUiOiJFY2hvIiwiaGFzaCI6IjRDRUM2NzNBN0RDQ0VBNkE0MTY1QkIxOTU4MzJDNzkzNjQ3MUNGN0FCNDUwMUY4MzdGOEQ2NzlGNDQwMEJDOTciLCJ0YWdzIjpbXSwiY2FwcyI6WyJ3YXNtY2xvdWQ6aHR0cHNlcnZlciJdLCJyZXYiOjQsInZlciI6IjAuMy40IiwicHJvdiI6ZmFsc2V9fQ.ZWyD6VQqzaYM1beD2x9Fdw4o_Bavy3ZG703Eg4cjhyJwUKLDUiVPVhqHFE6IXdV4cW6j93YbMT6VGq5iBDWmAg"
	t.Run("ParseWithClaims", func(t *testing.T) {
		_, err := jwt.ParseWithClaims(validJWT, &jwt.RegisteredClaims{}, KeyPairFromIssuer())
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("ComponentClaims", func(t *testing.T) {
		token, err := jwt.ParseWithClaims(validJWT, &WasCap{}, KeyPairFromIssuer())
		if err != nil {
			t.Fatal(err)
		}

		var componentClaims ComponentClaims
		wasCap := token.Claims.(*WasCap)
		err = wasCap.ParseCapability(&componentClaims)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestContext(t *testing.T) {
	raw := `{"application":{"policy":"","name":"appname"},"entity_jwt":"eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJxdmVOakZjcW51dWhQaVJUMkU1YWJXIiwiaWF0IjoxNzIxODM0ODg5LCJpc3MiOiJBQk9HQjRXNURPWDNVTzNSVldXUUdZU01WWEhSUFFZWFZaUDVVNFZGTUpEQ1lDV0FSN1M1Q1lNTyIsInN1YiI6Ik1DNUNDNFVENUxQRFo0QzdaTkFFQTRPWlEzQkVGTFNWUTc0MlczVEVUM09OS1M0RFJCVk5NNUlDIiwid2FzY2FwIjp7Im5hbWUiOiJodHRwLWhlbGxvLXdvcmxkIiwiaGFzaCI6IkNFOTAxOTJDOTlDMEIyQzYwOEIyRTJDQjYxOUE5MjUxRkI2ODE4NTZDMTU2ODFCMUJDRDYyRUVEQTJENTEyOEUiLCJ0YWdzIjpbIndhc21jbG91ZC5jb20vZXhwZXJpbWVudGFsIl0sInJldiI6MCwidmVyIjoiMC4xLjAiLCJwcm92IjpmYWxzZX0sIndhc2NhcF9yZXZpc2lvbiI6M30.8awbkvrBnRKLpz88s7GXYCW0onpKf_nNfsj7pXhCyvq8pm4y2IotrIPCdBvWqDvDouX4VAM6DQQUHuI-VdKYAA","host_jwt":"eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJuTGdta2Zud2p2Nkw1R28xSlNUdU0zIiwiaWF0IjoxNzIyMDE5OTk1LCJpc3MiOiJBQzNGU0IzT0VSQ1IzVU00WVNWUjJUQURFVlFWUTNITVpQQUtHS082QkNRSTRSNEFITFY2SVhSMiIsInN1YiI6Ik5ETlBUM0QzWVNUQzVKR0g2QVBKUDZBTVZYUVk2QklETVVXWkdTU1FXMjZWSjNINFBDRjJTU0ZSIiwid2FzY2FwIjp7Im5hbWUiOiJkZWxpY2F0ZS1icmVlemUtOTc4NSIsImxhYmVscyI6eyJzZWxmX3NpZ25lZCI6InRydWUifX0sIndhc2NhcF9yZXZpc2lvbiI6M30.5LM_GOpo-6qg0kDrIP_jswI_ZQfOILzHT-FHixvUeAf-1isamLg81S-rb84w6topfvevI6quyV3b-uHZt6q9BQ"}`
	ctx := &Context{}
	err := json.Unmarshal([]byte(raw), ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := ctx.IsValid(); err != nil {
		t.Fatal(err)
	}
}
