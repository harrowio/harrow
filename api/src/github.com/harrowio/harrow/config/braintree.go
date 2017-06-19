package config

import braintree "github.com/lionelbarrow/braintree-go"

type BraintreeConfig struct {
	Environment string
	MerchantId  string
	PublicKey   string
	PrivateKey  string
	CacheFile   string
}

func (c *Config) Braintree() *BraintreeConfig {
	return &BraintreeConfig{
		Environment: getEnvWithDefault("HAR_BRAINTREE_ENVIRONMENT", "sandbox"),
		MerchantId:  getEnvWithDefault("HAR_BRAINTREE_MERCHANT_ID", "9z379v9k6mp95dsm"),
		PublicKey:   getEnvWithDefault("HAR_BRAINTREE_PUBLIC_KEY", "59f3fjbs7tbknc7r"),
		PrivateKey:  getEnvWithDefault("HAR_BRAINTREE_PRIVATE_KEY", "688b56942e46e77c1d3b735b15ae78df"),
		CacheFile:   getEnvWithDefault("HAR_BRAINTREE_CACHE_FILE_PATH", "/tmp/braintree-cache-file.bin"),
	}
}

func (self *BraintreeConfig) NewClient() *braintree.Braintree {
	env, err := braintree.EnvironmentFromName(self.Environment)
	if err != nil {
		panic(err)
	}
	return braintree.New(env, self.MerchantId, self.PublicKey, self.PrivateKey)
}
