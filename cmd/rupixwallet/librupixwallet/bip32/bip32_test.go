package bip32

import (
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func TestBIP32SpecVectors(t *testing.T) {
	type testPath struct {
		path               string
		extendedPublicKey  string
		extendedPrivateKey string
	}

	type testVector struct {
		seed    string
		version [4]byte
		paths   []testPath
	}

	// test vectors are copied from https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#Test_Vectors
	testVectors := []testVector{
		{
			seed:    "000102030405060708090a0b0c0d0e0f",
			version: BitcoinMainnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
					extendedPrivateKey: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "xpub68Gmy5EdvgibQVfPdqkBBCHxA5htiqg55crXYuXoQRKfDBFA1WEjWgP6LHhwBZeNK1VTsfTFUHCdrfp1bgwQ9xv5ski8PX9rL2dZXvgGDnw",
					extendedPrivateKey: "xprv9uHRZZhk6KAJC1avXpDAp4MDc3sQKNxDiPvvkX8Br5ngLNv1TxvUxt4cV1rGL5hj6KCesnDYUhd7oWgT11eZG7XnxHrnYeSvkzY7d2bhkJ7",
				},
				{
					path:               "m/0'/1",
					extendedPublicKey:  "xpub6ASuArnXKPbfEwhqN6e3mwBcDTgzisQN1wXN9BJcM47sSikHjJf3UFHKkNAWbWMiGj7Wf5uMash7SyYq527Hqck2AxYysAA7xmALppuCkwQ",
					extendedPrivateKey: "xprv9wTYmMFdV23N2TdNG573QoEsfRrWKQgWeibmLntzniatZvR9BmLnvSxqu53Kw1UmYPxLgboyZQaXwTCg8MSY3H2EU4pWcQDnRnrVA1xe8fs",
				},
				{
					path:               "m/0'/1/2'",
					extendedPublicKey:  "xpub6D4BDPcP2GT577Vvch3R8wDkScZWzQzMMUm3PWbmWvVJrZwQY4VUNgqFJPMM3No2dFDFGTsxxpG5uJh7n7epu4trkrX7x7DogT5Uv6fcLW5",
					extendedPrivateKey: "xprv9z4pot5VBttmtdRTWfWQmoH1taj2axGVzFqSb8C9xaxKymcFzXBDptWmT7FwuEzG3ryjH4ktypQSAewRiNMjANTtpgP4mLTj34bhnZX7UiM",
				},
				{
					path:               "m/0'/1/2'/2",
					extendedPublicKey:  "xpub6FHa3pjLCk84BayeJxFW2SP4XRrFd1JYnxeLeU8EqN3vDfZmbqBqaGJAyiLjTAwm6ZLRQUMv1ZACTj37sR62cfN7fe5JnJ7dh8zL4fiyLHV",
					extendedPrivateKey: "xprvA2JDeKCSNNZky6uBCviVfJSKyQ1mDYahRjijr5idH2WwLsEd4Hsb2Tyh8RfQMuPh7f7RtyzTtdrbdqqsunu5Mm3wDvUAKRHSC34sJ7in334",
				},
				{
					path:               "m/0'/1/2'/2/1000000000",
					extendedPublicKey:  "xpub6H1LXWLaKsWFhvm6RVpEL9P4KfRZSW7abD2ttkWP3SSQvnyA8FSVqNTEcYFgJS2UaFcxupHiYkro49S8yGasTvXEYBVPamhGW6cFJodrTHy",
					extendedPrivateKey: "xprvA41z7zogVVwxVSgdKUHDy1SKmdb533PjDz7J6N6mV6uS3ze1ai8FHa8kmHScGpWmj4WggLyQjgPie1rFSruoUihUZREPSL39UNdE3BBDu76",
				},
			},
		},
		{
			seed:    "000102030405060708090a0b0c0d0e0f",
			version: KaspaMainnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "rpubcnbiN2K3n6qiL2X1KoomCkWRzFUxH831iVzQzajVQugfBDrC91dM2xD7whM7AP38qp4pxAGCNtRS8YE9wfYJSUw39MSRTJTQpbskYCrykUju",
					extendedPrivateKey: "rprv2uLskxh4r6hyqnBg2uSMVKK4bi6c5jUUoVs9ediX795neuUtyTq4Tm9orynQx9rzidNnbE95aZ2n9UgvkeQa9tZwuhc4fMPxxhifNo7bGJ2g",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "rpubcndyn22MFC6qKB8931ZdqdYY6aJWdW7VjBbqtYpsXssrTb8VKRrt85MUMUcfs8M3NpH7X2Vjod4WDWd85ca2pYpVsZ5DiDmK37ot6WWmHjFX",
					extendedPrivateKey: "rprv2uP9AxQNKBy6pvnok7CE8CMAi2vAS7YxpBUaYbouE7GywGmC9t4bYtJAGm3mHZaNgLRgHbSQSwPd2CSxUHzrLge2RmUjobMfrQjJEELbtQ9Z",
				},
				{
					path:               "m/0'/1",
					extendedPublicKey:  "rpubcng9uDou8aoiP1aBUjpXiEHRkdgVjW9E27vWj96eLpWefpfzT9fJS2vNath8SYHkimzjZovBujezh6vru5uCiEUKorH4ZhQKJkYQsoUhz9wa",
					extendedPrivateKey: "rprv2uRKJABvCafytmErBqT7zo64N6J9Y7ah77oFPC5g33un9WJhHbs1rqs4WB6xMAW9inWRyQFzt26aSLPUhRLeKToWsHFhXf7Si5XcbmDp2Rff",
				},
				{
					path:               "m/0'/1/2'",
					extendedPublicKey:  "rpubcnimBGLizHgZnsjyZzQw5bHTtrqNFmgp1TTkQPRwVzP27EXBZxR8rwMvWSiKGzAC38WqJRJAX7bZfZG1BnzkFHvUeSB2hnMNzUEL1thPnhUt",
					extendedPrivateKey: "rprv2uTvaCik4HYqJdQeH63XNA66WKT24P8H6TLV4SQyCDn9av9tQQcrHkJcRj9Ay8jfDHyTMziwoSWQLZbDT1MZWatxXdsG5p3gegoMpPoYh7vB",
				},
				{
					path:               "m/0'/1/2'/2",
					extendedPublicKey:  "rpubcnkza6mqwUAEmxDTHgg9AUndCweezQH8CtwdhePTyJpaibcow2BqE8wPS83JfPxLmbpxUZJeUALTn7gMBtJBT1WwuLxatcYGpUvEs3QR6SNw",
					extendedPrivateKey: "rprv2uW9y39s1U2WHht7znJjT3bFpQGJo1ibHtpNMhNVfYDiCHFWmUPYewt5MQTaRbQ4eMmb4ceBNMKrW2n7uCn6rnHYa37MBN8WMqmpyuJv2uE2",
				},
				{
					path:               "m/0'/1/2'/2/1000000000",
					extendedPublicKey:  "rpubcnniLaTTBbHcyUZEjoDhtnVdCjtEJDmwEhC2Ftfr7WtyDJkDKYc5tQ3YVksDcFDRV5XF24eaGhYANi6kCz9gHrn72DVzyR1rTHsrnHUwFrCc",
					extendedPrivateKey: "rprv2uXsjWqUFb9tVEDuStrJBMJFpCVt6qDQKh4kuwesokJ6gzNv9zooKCzER3KMdWKBiyAzKQ1AKCNPd2x8Gjr7auFC7Nc7QV3G587PLeUbyPLi",
				},
			},
		},
		{
			seed:    "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			version: BitcoinMainnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
					extendedPrivateKey: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
				},
				{
					path:               "m/0",
					extendedPublicKey:  "xpub69H7F5d8KSRgmmdJg2KhpAK8SR3DjMwAdkxj3ZuxV27CprR9LgpeyGmXUbC6wb7ERfvrnKZjXoUmmDznezpbZb7ap6r1D3tgFxHmwMkQTPH",
					extendedPrivateKey: "xprv9vHkqa6EV4sPZHYqZznhT2NPtPCjKuDKGY38FBWLvgaDx45zo9WQRUT3dKYnjwih2yJD9mkrocEZXo1ex8G81dwSM1fwqWpWkeS3v86pgKt",
				},
				{
					path:               "m/0/2147483647'",
					extendedPublicKey:  "xpub6ASAVgeehLbnwdqV6UKMHVzgqAG8Gr6riv3Fxxpj8ksbH9ebxaEyBLZ85ySDhKiLDBrQSARLq1uNRts8RuJiHjaDMBU4Zn9h8LZNnBC5y4a",
					extendedPrivateKey: "xprv9wSp6B7kry3Vj9m1zSnLvN3xH8RdsPP1Mh7fAaR7aRLcQMKTR2vidYEeEg2mUCTAwCd6vnxVrcjfy2kRgVsFawNzmjuHc2YmYRmagcEPdU9",
				},
				{
					path:               "m/0/2147483647'/1",
					extendedPublicKey:  "xpub6DF8uhdarytz3FWdA8TvFSvvAh8dP3283MY7p2V4SeE2wyWmG5mg5EwVvmdMVCQcoNJxGoWaU9DCWh89LojfZ537wTfunKau47EL2dhHKon",
					extendedPrivateKey: "xprv9zFnWC6h2cLgpmSA46vutJzBcfJ8yaJGg8cX1e5StJh45BBciYTRXSd25UEPVuesF9yog62tGAQtHjXajPPdbRCHuWS6T8XA2ECKADdw4Ef",
				},
				{
					path:               "m/0/2147483647'/1/2147483646'",
					extendedPublicKey:  "xpub6ERApfZwUNrhLCkDtcHTcxd75RbzS1ed54G1LkBUHQVHQKqhMkhgbmJbZRkrgZw4koxb5JaHWkY4ALHY2grBGRjaDMzQLcgJvLJuZZvRcEL",
					extendedPrivateKey: "xprvA1RpRA33e1JQ7ifknakTFpgNXPmW2YvmhqLQYMmrj4xJXXWYpDPS3xz7iAxn8L39njGVyuoseXzU6rcxFLJ8HFsTjSyQbLYnMpCqE2VbFWc",
				},
				{
					path:               "m/0/2147483647'/1/2147483646'/2",
					extendedPublicKey:  "xpub6FnCn6nSzZAw5Tw7cgR9bi15UV96gLZhjDstkXXxvCLsUXBGXPdSnLFbdpq8p9HmGsApME5hQTZ3emM2rnY5agb9rXpVGyy3bdW6EEgAtqt",
					extendedPrivateKey: "xprvA2nrNbFZABcdryreWet9Ea4LvTJcGsqrMzxHx98MMrotbir7yrKCEXw7nadnHM8Dq38EGfSh6dqA9QWTyefMLEcBYJUuekgW4BYPJcr9E7j",
				},
			},
		},
		{
			seed:    "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			version: KaspaMainnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "rpubcnbiN2K3n6qiL28WeuxqVCi9kVg6hofGYBaCisgEEt4n3WqDLBn7yP7HY3TC96HVoid4k4vrYHVwPG1tAnjDbcr36DNcv8wuUvFReUosfZnV",
					extendedPrivateKey: "rprv2uLskxh4r6hyqmoBN1bRmmWnMxHkWR6jdBSwNvfFw7TuXCTvAdyqQC3yTKrC1LNYosRY1yRY2q8pLaDergpxrVwGgA28uVjshGyHcAAK1x4u",
				},
				{
					path:               "m/0",
					extendedPublicKey:  "rpubcnez7J2jjarYQYQ6x3kDNGWZGrdqxWdkpjjx63VFgxUe1CofJm3U3XwrncvA2tNWEvwYuw9rHganMRBJrfsv1xShNVRMb3J3s3jYJv1J8Uwj",
					extendedPrivateKey: "rprv2uQ9WEQkoaiovJ4mf9NoeqKBtKFVm85Dujcgk6UHPBsmUtSN9DFBULtYhuMToySPeH5mqsRwmGJETvjHgF7TuSAS5ACYxtE3SQPCAXS9BHPs",
				},
				{
					path:               "m/0/2147483647'",
					extendedPublicKey:  "rpubcng9AYdmFxkiWiGK8UCD1jrEqFP4s47vWpu2cxtATcDQPf6tmNvtMk1ePEJQ9e77LiTUTazhtyoCx5rBCSnQ8gbA12VyeQ2Jsv7ouktGqVsN",
					extendedPrivateKey: "rprv2uRJZV1nKxcz2TvyqZpoJJesShziffZPbpmmH1sC9qcXsLjbbq8bnYxLJWhwnhh88BK6jeT9QKJjaMy2SyV531TsdavnJejmhCAXhHmaCVaD",
				},
				{
					path:               "m/0/2147483647'/1",
					extendedPublicKey:  "rpubcnix8xekC8Q1hoszGXrMahoB4auwNAJqn9LXUowpnv6kqKvkvgSR4dv2m56bHRyodJdw1Rdo8cvWnAeSDMgq5wvcucnBVcZk5qtUs1JETetY",
					extendedPrivateKey: "rprv2uU7Xu2mG8GHDZYeydUwsGbog3XbAmkJs9DG8rvrV9VtK1ZTm8e8VSrigMW9QjQKpVGTSPkDnirQngfoc2NbR1wgvihK7Vqk5fxxRmXuWeDi",
				},
				{
					path:               "m/0/2147483647'/1/2147483646'",
					extendedPublicKey:  "rpubcnk8AscgYjnyR6qDsGLB85JsFVeQjDHUHB3FNLfXCks25nH5rn7M5ASPrhkindML5G5aeE8rqfXqdpHbc3ZwbfHKMtgVzArqVi7ZSYFTRTVi",
					extendedPrivateKey: "rprv2uVHZozhcjfEvrVtaMxmQe7VrxG4XpiwNAuz2PeYtzG9ZTunhEK4VyP5mzCsoMpi72qk8hZzn7DzNVntyYKVuhnN6YdrRe3mi1YxwqJH2hov",
				},
				{
					path:               "m/0/2147483647'/1/2147483646'/2",
					extendedPublicKey:  "rpubcnmVCq3u4FyHer6QkzQJp44FDthwqTcPMqCsFkSshPesfrURRwkGqM1Lrn9o4kvgmn8nsW4NFZErdJif6sfdVyYAwXrL57E8EPQkdCsT9Gth",
					extendedPrivateKey: "rprv2uWebmRv8FqZAbm5U62u6crsqMKbe53rSq5buoRuPd419Y78GPwzG9x2n4cYoWqoB59brzKdbZKq4YLnVGds8km6pMVMvhTuRhvJVusVm3KM",
				},
			},
		},
		{
			seed:    "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be",
			version: BitcoinMainnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "xpub661MyMwAqRbcEZVB4dScxMAdx6d4nFc9nvyvH3v4gJL378CSRZiYmhRoP7mBy6gSPSCYk6SzXPTf3ND1cZAceL7SfJ1Z3GC8vBgp2epUt13",
					extendedPrivateKey: "xprv9s21ZrQH143K25QhxbucbDDuQ4naNntJRi4KUfWT7xo4EKsHt2QJDu7KXp1A3u7Bi1j8ph3EGsZ9Xvz9dGuVrtHHs7pXeTzjuxBrCmmhgC6",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "xpub68NZiKmJWnxxS6aaHmn81bvJeTESw724CRDs6HbuccFQN9Ku14VQrADWgqbhhTHBaohPX4CjNLf9fq9MYo6oDaPPLPxSb7gwQN3ih19Zm4Y",
					extendedPrivateKey: "xprv9uPDJpEQgRQfDcW7BkF7eTya6RPxXeJCqCJGHuCJ4GiRVLzkTXBAJMu2qaMWPrS7AANYqdq6vcBcBUdJCVVFceUvJFjaPdGZ2y9WACViL4L",
				},
			},
		},
		{
			seed:    "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be",
			version: KaspaMainnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "rpubcnbiN2K3n6qiL1BxpSMLHQhQnNKRoZXRotuyHGyFo9krqV5SbqvMwLNX4XSj7ut5SthpbtvjYgAmEhKX5dSG33BhELcX8sWMKhxwM1MAr7D2",
					extendedPrivateKey: "rprv2uLskxh4r6hyqkrdXXyvZyW3Ppw5cAxtttnhwKxHVP9zKAi9SJ85N9KCyoqvBHPn8x8CmYME8jZZ3vsGAvG7HHQmvgJhYhBDfZgwxp4f6qjp",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "rpubcne5ZmGsunD5gCj4DfVfnTxAT4g3BiNqiJQDE6Cwe64nCk6a4RR8oQqJmqAZdeEgC65KSftVHX7xjKnTRZgCDcRyB1iU2RMr8C9JFfh8kU1W",
					extendedPrivateKey: "rprv2uPExhetyn5MBxPivm8G52ko4XHgzKpJoJGwt9ByLKTugRjGtscrEDmzh7cGXdM74QGrBZJ21PJBWaQuKVUh33AyZ7ScbSLVUghucmQueXQJ",
				},
			},
		},
		{
			seed:    "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be",
			version: KaspaTestnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "rtubRMUvc4ZFRsPiMdpRNqx1q7NiRgdc4pQ5Ntv1hc2EMnw7JtQjkAGZ7D3ApbRH6M5HSn27KxMEzXAXorLHzvVWYBXfj7NYobStZC7csJ3DhaHF",
					extendedPrivateKey: "rtrvaHUpsBcaCfuzbM6Hw8qibPBfYGbbeb121AtjQgMPYT5hJersbdgWkBjUq79Mv3bhXxeAjWbNSgVkynYtEsfNp27uP3MYKTL5uNoAxYRtAZgq",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "rtubRMXHooX5ZYm5hqMWn56MLAdU6NzDSyFVHJQFeRFvCjF2g9RsCjmKyHVxXu97c5RtByPcAjJzjN7jJUoELrjSikmwfnUVh9JPMgHymxTrxSUV",
					extendedPrivateKey: "rtrvaHXC4vaQLMHMwYdPLMz46SSRCxxD2jrRuaNyMVb5PPPcfut14DBHcGCGYQuiGPZ2TQnp9XYAKLEPSS6XPSsxZmt71UVTNCVMiVp8cVma7s7s",
				},
			},
		},
		{
			seed:    "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be",
			version: KaspaDevnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "rdub75rMdYavDCtRhXfQUSTbw1egfvp3xdaxshzQSGPUnm1he9V7JjzqgXrErz2ZqPtPf6J4a28Ce1Z2QQYxTsv9wdQzxJgDQG2xt36QhLkG43CX",
					extendedPrivateKey: "rdrvNAyRMUa13MWB19AHxZ8vvy7K7FYHdvbxswSAvCNqXSPiQCXzup7YsiRSRUm3ES2DMQwDct3xpHwp8DpDyRtjsiTAfBqBvzHL92jHHArLWmFJ",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "rdub75tiqHYkLtFo3jCVsfbwS4uSLdAfLnSNn7UeP5dAdhKd1QWEmKVcYcK2aHkQM8EzQHfZQo5xNrWDu31topA68CfGtynAHotTgXGmc1Bs9xRT",
					extendedPrivateKey: "rdrvNB1nZDXqB2sYMLhPMnHGS2N4mwtu25TNnLvQs1cXNNhdmTZ8NPcKjntE8nXPamyYGs5s2tzkgwgSasMs817KdUDNHcy6yjSbx9kEw8FpE3aF",
				},
			},
		},
		{
			seed:    "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be",
			version: KaspaSimnetPrivate,
			paths: []testPath{
				{
					path:               "m",
					extendedPublicKey:  "rsub3j5At4CAdrNyTJFCMEXUazNtbvVmZfPeXXdVBakphQec3rNKaGj1S44FZUFzvuTvSg5ehqEyqSv6DGiqZU9q4yY8beT9Kxg1Hkd21DB3t82g",
					extendedPrivateKey: "rsrvCf559BFVQeuFh1X4uXRBMGBqiWTm9Rzb9ocCtf5yt4oC3cpTRk8y52kZZyz5kbzLXrhi7PV7HcFKPCwRoRKhLp8NFaS8qpZCdwJa6TZrrfPW",
				},
				{
					path:               "m/0'",
					extendedPublicKey:  "rsub3j7Y5o9zmXkLoVnHkTfp63deGcrNwpF4Rw7j8PzWYLxXR7PT2rDnJ8X3GmyqSdpXBsT9YcCjaHsHhuBmuQPmFYnQYKZ6DWXW6EoNusYB8uXy",
					extendedPrivateKey: "rsrvCf7SLvDKYLGd3D4AJkZWrKSbPCpNXar14D6SqUKfj177QsqatKdjw7DMHHkS6wwfTJrMXQRuAFywqrV4wzYH6ZtZt1a3tZiUT4KXkQsAAcaP",
				},
			},
		},
	}

	for i, vector := range testVectors {
		seed, err := hex.DecodeString(vector.seed)
		if err != nil {
			t.Fatalf("DecodeString: %+v", err)
		}

		masterKey, err := NewMaster(seed, vector.version)
		if err != nil {
			t.Fatalf("NewMaster: %+v", err)
		}

		for j, path := range vector.paths {
			extendedPrivateKey, err := masterKey.DeriveFromPath(path.path)
			if err != nil {
				t.Fatalf("Path: %+v", err)
			}

			if extendedPrivateKey.String() != path.extendedPrivateKey {
				t.Fatalf("Test (%d, %d): expected extended private key %s but got %s", i, j, path.extendedPrivateKey, extendedPrivateKey.String())
			}

			decodedExtendedPrivateKey, err := DeserializeExtendedKey(extendedPrivateKey.String())
			if err != nil {
				t.Fatalf("DeserializeExtendedKey: %+v", err)
			}

			if extendedPrivateKey.String() != decodedExtendedPrivateKey.String() {
				t.Fatalf("Test (%d, %d): deserializing and serializing the extended private key didn't preserve the data", i, j)
			}

			extendedPublicKey, err := extendedPrivateKey.Public()
			if err != nil {
				t.Fatalf("Public: %+v", err)
			}

			if extendedPublicKey.String() != path.extendedPublicKey {
				t.Fatalf("Test (%d, %d): expected extended public key %s but got %s", i, j, path.extendedPublicKey, extendedPublicKey.String())
			}

			decodedExtendedPublicKey, err := DeserializeExtendedKey(extendedPublicKey.String())
			if err != nil {
				t.Fatalf("DeserializeExtendedPublicKey: %+v", err)
			}

			if extendedPublicKey.String() != decodedExtendedPublicKey.String() {
				t.Fatalf("Test (%d, %d): deserializing and serializing the ext pub didn't preserve the data", i, j)
			}
		}
	}
}

// TestExtendedKey_DeriveFromPath checks that path that derive from extended private key and extended
// public key lead to the same public keys.
func TestExtendedKey_DeriveFromPath(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	seed, err := GenerateSeed()
	if err != nil {
		t.Fatalf("GenerateSeed: %+v", err)
	}

	master, err := NewMaster(seed, KaspaMainnetPrivate)
	if err != nil {
		t.Fatalf("GenerateSeed: %+v", err)
	}

	masterPublic, err := master.Public()
	if err != nil {
		t.Fatalf("Public: %+v", err)
	}

	for i := 0; i < 10; i++ {
		numIndexes := 1 + r.Intn(100)
		indexes := make([]string, numIndexes)
		for i := 0; i < numIndexes; i++ {
			index := r.Intn(hardenedIndexStart)
			indexes[i] = strconv.Itoa(int(index))
		}

		indexesStr := strings.Join(indexes, "/")
		pathPrivate := "m/" + indexesStr
		pathPublic := "M/" + indexesStr

		extendedPrivateKey, err := master.DeriveFromPath(pathPrivate)
		if err != nil {
			t.Fatalf("Path: %+v", err)
		}

		extendedPublicKeyFromPrivateKey, err := extendedPrivateKey.Public()
		if err != nil {
			t.Fatalf("Public: %+v", err)
		}

		extendedPublicKey, err := masterPublic.DeriveFromPath(pathPublic)
		if err != nil {
			t.Fatalf("Path: %+v", err)
		}

		if extendedPublicKeyFromPrivateKey.String() != extendedPublicKey.String() {
			t.Fatalf("Path gives different result from private and public master keys")
		}
	}
}

// TestPublicParentPublicChildDerivation was copied and modified from https://github.com/tyler-smith/go-bip32/blob/master/bip32_test.go
func TestPublicParentPublicChildDerivation(t *testing.T) {
	// Generated using https://iancoleman.github.io/bip39/
	// Root key:
	// xprv9s21ZrQH143K2Cfj4mDZBcEecBmJmawReGwwoAou2zZzG45bM6cFPJSvobVTCB55L6Ld2y8RzC61CpvadeAnhws3CHsMFhNjozBKGNgucYm
	// Derivation Path m/44'/60'/0'/0:
	// xprv9zy5o7z1GMmYdaeQdmabWFhUf52Ytbpe3G5hduA4SghboqWe7aDGWseN8BJy1GU72wPjkCbBE1hvbXYqpCecAYdaivxjNnBoSNxwYD4wHpW
	// xpub6DxSCdWu6jKqr4isjo7bsPeDD6s3J4YVQV1JSHZg12Eagdqnf7XX4fxqyW2sLhUoFWutL7tAELU2LiGZrEXtjVbvYptvTX5Eoa4Mamdjm9u
	extendedMasterPublic, err := DeserializeExtendedKey("xpub6DxSCdWu6jKqr4isjo7bsPeDD6s3J4YVQV1JSHZg12Eagdqnf7XX4fxqyW2sLhUoFWutL7tAELU2LiGZrEXtjVbvYptvTX5Eoa4Mamdjm9u")
	if err != nil {
		t.Fatalf("DeserializeExtendedPublicKey: %+v", err)
	}

	type testChildKey struct {
		pathFragment uint32
		privKey      string
		pubKey       string
		hexPubKey    string
	}

	expectedChildren := []testChildKey{
		{pathFragment: 0, hexPubKey: "0243187e1a2ba9ba824f5f81090650c8f4faa82b7baf93060d10b81f4b705afd46"},
		{pathFragment: 1, hexPubKey: "023790d11eb715c4320d8e31fba3a09b700051dc2cdbcce03f44b11c274d1e220b"},
		{pathFragment: 2, hexPubKey: "0302c5749c3c75cea234878ae3f4d8f65b75d584bcd7ed0943b016d6f6b59a2bad"},
		{pathFragment: 3, hexPubKey: "03f0440c94e5b14ea5b15875934597afff541bec287c6e65dc1102cafc07f69699"},
		{pathFragment: 4, hexPubKey: "026419d0d8996707605508ac44c5871edc7fe206a79ef615b74f2eea09c5852e2b"},
		{pathFragment: 5, hexPubKey: "02f63c6f195eea98bdb163c4a094260dea71d264b21234bed4df3899236e6c2298"},
		{pathFragment: 6, hexPubKey: "02d74709cd522081064858f393d009ead5a0ecd43ede3a1f57befcc942025cb5f9"},
		{pathFragment: 7, hexPubKey: "03e54bb92630c943d38bbd8a4a2e65fca7605e672d30a0e545a7198cbb60729ceb"},
		{pathFragment: 8, hexPubKey: "027e9d5acd14d39c4938697fba388cd2e8f31fc1c5dc02fafb93a10a280de85199"},
		{pathFragment: 9, hexPubKey: "02a167a9f0d57468fb6abf2f3f7967e2cadf574314753a06a9ef29bc76c54638d2"},

		{pathFragment: 100, hexPubKey: "020db9ba00ddf68428e3f5bfe54252bbcd75b21e42f51bf3bfc4172bf0e5fa7905"},
		{pathFragment: 101, hexPubKey: "0299e3790956570737d6164e6fcda5a3daa304065ca95ba46bc73d436b84f34d46"},
		{pathFragment: 102, hexPubKey: "0202e0732c4c5d2b1036af173640e01957998cfd4f9cdaefab6ffe76eb869e2c59"},
		{pathFragment: 103, hexPubKey: "03d050adbd996c0c5d737ff638402dfbb8c08e451fef10e6d62fb57887c1ac6cb2"},
		{pathFragment: 104, hexPubKey: "038d466399e2d68b4b16043ad4d88893b3b2f84fc443368729a973df1e66f4f530"},
		{pathFragment: 105, hexPubKey: "034811e2f0c8c50440c08c2c9799b99c911c036e877e8325386ff61723ae3ffdce"},
		{pathFragment: 106, hexPubKey: "026339fd5842921888e711a6ba9104a5f0c94cc0569855273cf5faefdfbcd3cc29"},
		{pathFragment: 107, hexPubKey: "02833705c1069fab2aa92c6b0dac27807290d72e9f52378d493ac44849ca003b22"},
		{pathFragment: 108, hexPubKey: "032d2639bde1eb7bdf8444bd4f6cc26a9d1bdecd8ea15fac3b992c3da68d9d1df5"},
		{pathFragment: 109, hexPubKey: "02479c6d4a64b93a2f4343aa862c938fbc658c99219dd7bebb4830307cbd76c9e9"},
	}

	for i, child := range expectedChildren {
		extendedPublicKey, err := extendedMasterPublic.Child(child.pathFragment)
		if err != nil {
			t.Fatalf("Child: %+v", err)
		}

		publicKey, err := extendedPublicKey.PublicKey()
		if err != nil {
			t.Fatalf("PublicKey: %+v", err)
		}

		pubKeyBytes, err := publicKey.Serialize()
		if err != nil {
			t.Fatalf("Serialize: %+v", err)
		}

		pubKeyHex := hex.EncodeToString(pubKeyBytes[:])
		if child.hexPubKey != pubKeyHex {
			t.Fatalf("Test #%d: expected public key %s but got %s", i, child.hexPubKey, pubKeyHex)
		}
	}
}
