# Paper font assets

Paper ships static, normal-style WOFF2 faces for Merriweather 700 and Open Sans
400, 600, and 700. They are generated from variable TrueType sources in the
Google Fonts repository at immutable commit
`2d85e20401920891efb7cd6272d6339685df2820` (retrieved 2026-08-08).

| Input | SHA-256 |
| --- | --- |
| `ofl/merriweather/Merriweather[opsz,wdth,wght].ttf` | `d0ed0e359e396af7ad05e73dffd11a3a4c326ea0d0283c56bd9361cb2cc86a96` |
| `ofl/opensans/OpenSans[wdth,wght].ttf` | `36643644f318a812aab2d2ed3bb98f8cf0872527f835fe9398d95fe6b9adb878` |

Merriweather is instantiated at `opsz=18`, `wdth=100`, `wght=700`. Open Sans
is instantiated at `wdth=100` and each shipped weight. FontTools 4.59.0 with
its WOFF extra performs the conversion; `--no-recalc-timestamp` keeps output
deterministic.

Reproduction outline (run from this directory, using a temporary Python
environment):

```sh
font_commit=2d85e20401920891efb7cd6272d6339685df2820
curl -fsSL "https://raw.githubusercontent.com/google/fonts/$font_commit/ofl/merriweather/Merriweather%5Bopsz%2Cwdth%2Cwght%5D.ttf" -o merriweather-variable.ttf
curl -fsSL "https://raw.githubusercontent.com/google/fonts/$font_commit/ofl/opensans/OpenSans%5Bwdth%2Cwght%5D.ttf" -o open-sans-variable.ttf
curl -fsSL "https://raw.githubusercontent.com/google/fonts/$font_commit/ofl/merriweather/OFL.txt" -o OFL-Merriweather.txt
curl -fsSL "https://raw.githubusercontent.com/google/fonts/$font_commit/ofl/opensans/OFL.txt" -o OFL-Open-Sans.txt
python3 -m pip install 'fonttools[woff]==4.59.0'
fonttools varLib.instancer merriweather-variable.ttf opsz=18 wdth=100 wght=700 --no-recalc-timestamp -o merriweather-700.ttf
fonttools varLib.instancer open-sans-variable.ttf wdth=100 wght=400 --no-recalc-timestamp -o open-sans-400.ttf
fonttools varLib.instancer open-sans-variable.ttf wdth=100 wght=600 --no-recalc-timestamp -o open-sans-600.ttf
fonttools varLib.instancer open-sans-variable.ttf wdth=100 wght=700 --no-recalc-timestamp -o open-sans-700.ttf
fonttools ttLib.woff2 compress merriweather-700.ttf -o merriweather-700.woff2
fonttools ttLib.woff2 compress open-sans-400.ttf -o open-sans-400.woff2
fonttools ttLib.woff2 compress open-sans-600.ttf -o open-sans-600.woff2
fonttools ttLib.woff2 compress open-sans-700.ttf -o open-sans-700.woff2
```

Upstream sources:

- <https://github.com/google/fonts/blob/2d85e20401920891efb7cd6272d6339685df2820/ofl/merriweather/Merriweather%5Bopsz%2Cwdth%2Cwght%5D.ttf>
- <https://github.com/google/fonts/blob/2d85e20401920891efb7cd6272d6339685df2820/ofl/opensans/OpenSans%5Bwdth%2Cwght%5D.ttf>

Both families use the SIL Open Font License 1.1. The unmodified upstream
license texts are included as `OFL-Merriweather.txt` and `OFL-Open-Sans.txt`.
Output hashes are recorded in the parent asset manifest.
