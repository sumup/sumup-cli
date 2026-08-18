# Changelog

## [0.4.0](https://github.com/sumup/sumup-cli/compare/v0.3.1...v0.4.0) (2026-08-18)


### Features

* autocomplete, man pages ([#120](https://github.com/sumup/sumup-cli/issues/120)) ([a03c6b6](https://github.com/sumup/sumup-cli/commit/a03c6b637a334c63ec7c6a7aaf542bc5d56c8b67))
* **build:** more release targets ([#17](https://github.com/sumup/sumup-cli/issues/17)) ([84a8c4d](https://github.com/sumup/sumup-cli/commit/84a8c4df29f7dd37bda3affd3e0957a6397f95f0))
* **cd:** publish homebrew cask ([#121](https://github.com/sumup/sumup-cli/issues/121)) ([cc43546](https://github.com/sumup/sumup-cli/commit/cc435469dfe364c1b446c2a813eaa0a2ec0b2225))
* **cli:** expand sdk command coverage ([#38](https://github.com/sumup/sumup-cli/issues/38)) ([5f2aa0b](https://github.com/sumup/sumup-cli/commit/5f2aa0b02879c887e4f9a404898b73c4a7b01dc8))
* **cmd:** list roles ([#70](https://github.com/sumup/sumup-cli/issues/70)) ([1637521](https://github.com/sumup/sumup-cli/commit/1637521677902eafdb041c90336558279e24349b))
* **commands:** standardize mutation acknowledgements ([#56](https://github.com/sumup/sumup-cli/issues/56)) ([6c94384](https://github.com/sumup/sumup-cli/commit/6c9438480865283fbff54bc33208d02ea16c1b5d))
* **deps:** bump all to latest ([#25](https://github.com/sumup/sumup-cli/issues/25)) ([b61b2cb](https://github.com/sumup/sumup-cli/commit/b61b2cba4468c706762f7136226940f0082dc4dc))
* **display:** add configurable table options ([#53](https://github.com/sumup/sumup-cli/issues/53)) ([b43876b](https://github.com/sumup/sumup-cli/commit/b43876bbf2e59056f46e5b253e6708ecdbc565f5))
* **display:** add sectioned detail rendering ([#51](https://github.com/sumup/sumup-cli/issues/51)) ([0721c98](https://github.com/sumup/sumup-cli/commit/0721c98965981dc20618543efdb9621dd61fbfa7))
* **display:** add shared mutation renderer ([#67](https://github.com/sumup/sumup-cli/issues/67)) ([b8276de](https://github.com/sumup/sumup-cli/commit/b8276de14d9f59ebe37335219a3f30a1d6e35d44))
* init ([d34c88b](https://github.com/sumup/sumup-cli/commit/d34c88b858752fa19bc62ab0dd56918a1d0d4a61))
* **members:** get, update support ([#27](https://github.com/sumup/sumup-cli/issues/27)) ([5ec0eab](https://github.com/sumup/sumup-cli/commit/5ec0eab5a7db26d95fd4dcb033ec1d76696b18c5))
* openapi endpoints parity ([#117](https://github.com/sumup/sumup-cli/issues/117)) ([4c866f7](https://github.com/sumup/sumup-cli/commit/4c866f7d787293ec63bbcc619a57e9af8cea6091))
* release please setup ([#19](https://github.com/sumup/sumup-cli/issues/19)) ([18c8a35](https://github.com/sumup/sumup-cli/commit/18c8a353e6f293a6001edf96e90dc1b0bf3e5ebd))
* release setup with goreleaser ([#16](https://github.com/sumup/sumup-cli/issues/16)) ([c202f62](https://github.com/sumup/sumup-cli/commit/c202f622933468e0f2a1306a39cbb73b79b437a7))
* support customers ([#26](https://github.com/sumup/sumup-cli/issues/26)) ([05c7a77](https://github.com/sumup/sumup-cli/commit/05c7a77a7f2545f00d3684b6d6e3dc685fceebf4))
* **tooling:** make install target ([#55](https://github.com/sumup/sumup-cli/issues/55)) ([b3d9fbc](https://github.com/sumup/sumup-cli/commit/b3d9fbc32eb189a6c677a483c739e7e69ae72c50))
* update checkout cmd ([#110](https://github.com/sumup/sumup-cli/issues/110)) ([de954b3](https://github.com/sumup/sumup-cli/commit/de954b3823594cf5e9d2cdc5aaf795c52eecffa3))
* update SumUp Go SDK ([#14](https://github.com/sumup/sumup-cli/issues/14)) ([f135135](https://github.com/sumup/sumup-cli/commit/f1351355d6b16d31acf9c6c7649a44bb87dcebc8))
* update testing setup ([#41](https://github.com/sumup/sumup-cli/issues/41)) ([ba39555](https://github.com/sumup/sumup-cli/commit/ba395558342281a3c2710f890b8853a4de7279b5))
* use exactly one command per operation ([#118](https://github.com/sumup/sumup-cli/issues/118)) ([495d37f](https://github.com/sumup/sumup-cli/commit/495d37fceaded241b8406a0a2d9e626a79649202))
* **version:** honor shared output settings ([#49](https://github.com/sumup/sumup-cli/issues/49)) ([0c6fc93](https://github.com/sumup/sumup-cli/commit/0c6fc9304971753238f97f55ae266b33a4db4622))


### Bug Fixes

* **cd:** publish releases via draft for immutable GitHub releases ([#124](https://github.com/sumup/sumup-cli/issues/124)) ([8d2684e](https://github.com/sumup/sumup-cli/commit/8d2684ed06f3582ecc90a9589303f7fd6530732b))
* **cd:** simplify release flow without immutable drafts ([#126](https://github.com/sumup/sumup-cli/issues/126)) ([46394ba](https://github.com/sumup/sumup-cli/commit/46394ba57ec125a8afc025632cc9e669917483e0))
* **cli:** honor merchant context fallback ([#40](https://github.com/sumup/sumup-cli/issues/40)) ([b4bfee8](https://github.com/sumup/sumup-cli/commit/b4bfee844c3858185e4e771b672b6332016a7b7c))
* **commands:** handle nil list responses safely ([#59](https://github.com/sumup/sumup-cli/issues/59)) ([902f838](https://github.com/sumup/sumup-cli/commit/902f838ea9e523611cd2ffe32bb246a4da044311))
* **commands:** honor exact timestamp formatting consistently ([#62](https://github.com/sumup/sumup-cli/issues/62)) ([15b78bd](https://github.com/sumup/sumup-cli/commit/15b78bdd255e8e595fa1438a73a5df753e52ae53))
* **context:** ignore stale membership search results ([#48](https://github.com/sumup/sumup-cli/issues/48)) ([8c1d316](https://github.com/sumup/sumup-cli/commit/8c1d316bf60e958b600567b669ce9cf934b02f1b))
* **message:** respect no-color terminal settings ([#64](https://github.com/sumup/sumup-cli/issues/64)) ([e6e3187](https://github.com/sumup/sumup-cli/commit/e6e3187b17c39124e673d3433949a7ace3d74eb2))
* parse monetary CLI amounts as decimals ([#45](https://github.com/sumup/sumup-cli/issues/45)) ([ef0e886](https://github.com/sumup/sumup-cli/commit/ef0e886a948517c2a45d515598928e14183d9d6f))
* **readers:** avoid panic on sparse problem responses ([#60](https://github.com/sumup/sumup-cli/issues/60)) ([df6f67d](https://github.com/sumup/sumup-cli/commit/df6f67d431e2ab51c80446a0749b78fed44b14eb))

## [0.3.0](https://github.com/sumup/sumup-cli/compare/v0.2.0...v0.3.0) (2026-08-18)


### Features

* autocomplete, man pages ([#120](https://github.com/sumup/sumup-cli/issues/120)) ([a03c6b6](https://github.com/sumup/sumup-cli/commit/a03c6b637a334c63ec7c6a7aaf542bc5d56c8b67))
* **cd:** publish homebrew cask ([#121](https://github.com/sumup/sumup-cli/issues/121)) ([cc43546](https://github.com/sumup/sumup-cli/commit/cc435469dfe364c1b446c2a813eaa0a2ec0b2225))

## [0.2.0](https://github.com/sumup/sumup-cli/compare/v0.1.0...v0.2.0) (2026-08-15)


### Features

* **build:** more release targets ([#17](https://github.com/sumup/sumup-cli/issues/17)) ([84a8c4d](https://github.com/sumup/sumup-cli/commit/84a8c4df29f7dd37bda3affd3e0957a6397f95f0))
* **cli:** expand sdk command coverage ([#38](https://github.com/sumup/sumup-cli/issues/38)) ([5f2aa0b](https://github.com/sumup/sumup-cli/commit/5f2aa0b02879c887e4f9a404898b73c4a7b01dc8))
* **cmd:** list roles ([#70](https://github.com/sumup/sumup-cli/issues/70)) ([1637521](https://github.com/sumup/sumup-cli/commit/1637521677902eafdb041c90336558279e24349b))
* **commands:** standardize mutation acknowledgements ([#56](https://github.com/sumup/sumup-cli/issues/56)) ([6c94384](https://github.com/sumup/sumup-cli/commit/6c9438480865283fbff54bc33208d02ea16c1b5d))
* **deps:** bump all to latest ([#25](https://github.com/sumup/sumup-cli/issues/25)) ([b61b2cb](https://github.com/sumup/sumup-cli/commit/b61b2cba4468c706762f7136226940f0082dc4dc))
* **display:** add configurable table options ([#53](https://github.com/sumup/sumup-cli/issues/53)) ([b43876b](https://github.com/sumup/sumup-cli/commit/b43876bbf2e59056f46e5b253e6708ecdbc565f5))
* **display:** add sectioned detail rendering ([#51](https://github.com/sumup/sumup-cli/issues/51)) ([0721c98](https://github.com/sumup/sumup-cli/commit/0721c98965981dc20618543efdb9621dd61fbfa7))
* **display:** add shared mutation renderer ([#67](https://github.com/sumup/sumup-cli/issues/67)) ([b8276de](https://github.com/sumup/sumup-cli/commit/b8276de14d9f59ebe37335219a3f30a1d6e35d44))
* init ([d34c88b](https://github.com/sumup/sumup-cli/commit/d34c88b858752fa19bc62ab0dd56918a1d0d4a61))
* **members:** get, update support ([#27](https://github.com/sumup/sumup-cli/issues/27)) ([5ec0eab](https://github.com/sumup/sumup-cli/commit/5ec0eab5a7db26d95fd4dcb033ec1d76696b18c5))
* openapi endpoints parity ([#117](https://github.com/sumup/sumup-cli/issues/117)) ([4c866f7](https://github.com/sumup/sumup-cli/commit/4c866f7d787293ec63bbcc619a57e9af8cea6091))
* release please setup ([#19](https://github.com/sumup/sumup-cli/issues/19)) ([18c8a35](https://github.com/sumup/sumup-cli/commit/18c8a353e6f293a6001edf96e90dc1b0bf3e5ebd))
* release setup with goreleaser ([#16](https://github.com/sumup/sumup-cli/issues/16)) ([c202f62](https://github.com/sumup/sumup-cli/commit/c202f622933468e0f2a1306a39cbb73b79b437a7))
* support customers ([#26](https://github.com/sumup/sumup-cli/issues/26)) ([05c7a77](https://github.com/sumup/sumup-cli/commit/05c7a77a7f2545f00d3684b6d6e3dc685fceebf4))
* **tooling:** make install target ([#55](https://github.com/sumup/sumup-cli/issues/55)) ([b3d9fbc](https://github.com/sumup/sumup-cli/commit/b3d9fbc32eb189a6c677a483c739e7e69ae72c50))
* update checkout cmd ([#110](https://github.com/sumup/sumup-cli/issues/110)) ([de954b3](https://github.com/sumup/sumup-cli/commit/de954b3823594cf5e9d2cdc5aaf795c52eecffa3))
* update SumUp Go SDK ([#14](https://github.com/sumup/sumup-cli/issues/14)) ([f135135](https://github.com/sumup/sumup-cli/commit/f1351355d6b16d31acf9c6c7649a44bb87dcebc8))
* update testing setup ([#41](https://github.com/sumup/sumup-cli/issues/41)) ([ba39555](https://github.com/sumup/sumup-cli/commit/ba395558342281a3c2710f890b8853a4de7279b5))
* use exactly one command per operation ([#118](https://github.com/sumup/sumup-cli/issues/118)) ([495d37f](https://github.com/sumup/sumup-cli/commit/495d37fceaded241b8406a0a2d9e626a79649202))
* **version:** honor shared output settings ([#49](https://github.com/sumup/sumup-cli/issues/49)) ([0c6fc93](https://github.com/sumup/sumup-cli/commit/0c6fc9304971753238f97f55ae266b33a4db4622))


### Bug Fixes

* **cli:** honor merchant context fallback ([#40](https://github.com/sumup/sumup-cli/issues/40)) ([b4bfee8](https://github.com/sumup/sumup-cli/commit/b4bfee844c3858185e4e771b672b6332016a7b7c))
* **commands:** handle nil list responses safely ([#59](https://github.com/sumup/sumup-cli/issues/59)) ([902f838](https://github.com/sumup/sumup-cli/commit/902f838ea9e523611cd2ffe32bb246a4da044311))
* **commands:** honor exact timestamp formatting consistently ([#62](https://github.com/sumup/sumup-cli/issues/62)) ([15b78bd](https://github.com/sumup/sumup-cli/commit/15b78bdd255e8e595fa1438a73a5df753e52ae53))
* **context:** ignore stale membership search results ([#48](https://github.com/sumup/sumup-cli/issues/48)) ([8c1d316](https://github.com/sumup/sumup-cli/commit/8c1d316bf60e958b600567b669ce9cf934b02f1b))
* **message:** respect no-color terminal settings ([#64](https://github.com/sumup/sumup-cli/issues/64)) ([e6e3187](https://github.com/sumup/sumup-cli/commit/e6e3187b17c39124e673d3433949a7ace3d74eb2))
* parse monetary CLI amounts as decimals ([#45](https://github.com/sumup/sumup-cli/issues/45)) ([ef0e886](https://github.com/sumup/sumup-cli/commit/ef0e886a948517c2a45d515598928e14183d9d6f))
* **readers:** avoid panic on sparse problem responses ([#60](https://github.com/sumup/sumup-cli/issues/60)) ([df6f67d](https://github.com/sumup/sumup-cli/commit/df6f67d431e2ab51c80446a0749b78fed44b14eb))
