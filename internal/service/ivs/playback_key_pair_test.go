package ivs_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfivs "github.com/hashicorp/terraform-provider-aws/internal/service/ivs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// There is an AWS maximum service quota of 3 playback key pairs, so tests
// cannot be ran in parallel. The key **MUST** use ECDSA which is not provided
// by acctest, so crypto functions are used in this test file. Furthermore,
// multiple playback key pairs cannot use the same public key, so using a static
// test fixture file is discouraged.

func testAccPlaybackKeyPair_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var playbackKeyPair ivs.PlaybackKeyPair
	resourceName := "aws_ivs_playback_key_pair.test"
	privateKey := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
	publicKeyPEM, fingerprint := acctest.TLSECDSAPublicKeyPEM(t, privateKey)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccPlaybackKeyPairPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaybackKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaybackKeyPairConfig_basic(publicKeyPEM),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &playbackKeyPair),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", fingerprint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ivs", regexp.MustCompile(`playback-key/.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
		},
	})
}

func testAccPlaybackKeyPair_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ivs.PlaybackKeyPair
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_playback_key_pair.test"
	privateKey1 := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
	publicKeyPEM1, fingerprint1 := acctest.TLSECDSAPublicKeyPEM(t, privateKey1)
	privateKey2 := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
	publicKeyPEM2, fingerprint2 := acctest.TLSECDSAPublicKeyPEM(t, privateKey2)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccPlaybackKeyPairPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaybackKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaybackKeyPairConfig_name(rName1, publicKeyPEM1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", fingerprint1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccPlaybackKeyPairConfig_name(rName2, publicKeyPEM2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &v2),
					testAccCheckPlaybackKeyPairRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", fingerprint2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccPlaybackKeyPair_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 ivs.PlaybackKeyPair
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivs_playback_key_pair.test"
	privateKey := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
	publicKeyPEM, _ := acctest.TLSECDSAPublicKeyPEM(t, privateKey)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccPlaybackKeyPairPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaybackKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaybackKeyPairConfig_tags1(rName, publicKeyPEM, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
			{
				Config: testAccPlaybackKeyPairConfig_tags2(rName, publicKeyPEM, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPlaybackKeyPairConfig_tags1(rName, publicKeyPEM, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccPlaybackKeyPair_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var playbackkeypair ivs.PlaybackKeyPair
	resourceName := "aws_ivs_playback_key_pair.test"
	privateKey := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
	publicKey, _ := acctest.TLSECDSAPublicKeyPEM(t, privateKey)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, ivs.EndpointsID)
			testAccPlaybackKeyPairPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ivs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaybackKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaybackKeyPairConfig_basic(publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaybackKeyPairExists(ctx, resourceName, &playbackkeypair),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfivs.ResourcePlaybackKeyPair(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPlaybackKeyPairDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ivs_playback_key_pair" {
				continue
			}

			input := &ivs.GetPlaybackKeyPairInput{
				Arn: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetPlaybackKeyPairWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			return create.Error(names.IVS, create.ErrActionCheckingDestroyed, tfivs.ResNamePlaybackKeyPair, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPlaybackKeyPairExists(ctx context.Context, name string, playbackkeypair *ivs.PlaybackKeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNamePlaybackKeyPair, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNamePlaybackKeyPair, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

		resp, err := tfivs.FindPlaybackKeyPairByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.IVS, create.ErrActionCheckingExistence, tfivs.ResNamePlaybackKeyPair, rs.Primary.ID, err)
		}

		*playbackkeypair = *resp

		return nil
	}
}

func testAccPlaybackKeyPairPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IVSConn(ctx)

	input := &ivs.ListPlaybackKeyPairsInput{}
	_, err := conn.ListPlaybackKeyPairsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPlaybackKeyPairRecreated(before, after *ivs.PlaybackKeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before == after {
			return fmt.Errorf("Expected Playback Key Pair IDs to change, %s", before)
		}

		return nil
	}
}

func testAccPlaybackKeyPairConfig_basic(publicKey string) string {
	// acctest.TLSPEMEscapeNewlines is not necessary for publicKey
	return fmt.Sprintf(`
resource aws_ivs_playback_key_pair "test" {
	public_key = %[1]q
}
`, publicKey)
}

func testAccPlaybackKeyPairConfig_name(rName, publicKey string) string {
	// acctest.TLSPEMEscapeNewlines is not necessary for publicKey
	return fmt.Sprintf(`
resource aws_ivs_playback_key_pair "test" {
	name = %[1]q
	public_key = %[2]q
}
`, rName, publicKey)
}

func testAccPlaybackKeyPairConfig_tags1(rName, publicKey, tagKey1, tagValue1 string) string {
	// acctest.TLSPEMEscapeNewlines is not necessary for publicKey
	return fmt.Sprintf(`
resource aws_ivs_playback_key_pair "test" {
	name = %[1]q
	public_key = %[2]q
	tags = {
		%[3]q = %[4]q
	}
}
`, rName, publicKey, tagKey1, tagValue1)
}

func testAccPlaybackKeyPairConfig_tags2(rName, publicKey, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	// acctest.TLSPEMEscapeNewlines is not necessary for publicKey
	return fmt.Sprintf(`
resource aws_ivs_playback_key_pair "test" {
	name = %[1]q
	public_key = %[2]q
	tags = {
		%[3]q = %[4]q
		%[5]q = %[6]q
	}
}
`, rName, publicKey, tagKey1, tagValue1, tagKey2, tagValue2)
}
