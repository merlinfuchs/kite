---
sidebar_position: 37.83
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Look up with Popcat

<EmbedFlowNode type="action_popcat_search" />

The `Look up with Popcat` block looks something up, like the weather of a place, a GitHub user, a movie, a song or a country. Enable Popcat under [Integrations](../../integrations.md) first. It needs no key, but it's off until you turn it on, as its blocks send what you enter to [Popcat](https://popcat.xyz/api), a free API that isn't run by Kite.

Pick what to look up and enter what to search for. The result is what Popcat found, with the fields listed for it, like `{{ result('lookup').followers }}` for a GitHub user. The weather is a list of matching places, so `{{ result('lookup')[0].current.temperature }}` is the temperature at the first one. If Popcat finds nothing, the block fails with Popcat's error, which an [error handler](../controls/control_error_handler.md) can catch.

| Look Up | Settings | What it returns |
| --- | --- | --- |
| Weather | Place | The weather of a place. Returns a list of matching places, each with location.name, location.degreetype, current.temperature, current.feelslike, current.skytext, current.humidity, current.winddisplay, current.imageUrl and forecast, a list of days with day, low, high, skytextday and precip. |
| GitHub user | Username | A GitHub user. Returns url, avatar, name, bio, company, location, blog, followers, following, public_repos and created_at. |
| Steam game | Game | A game on Steam. Returns name, description, price, thumbnail, banner, website, developers and publishers. |
| Movie or show | Title | A movie or show on IMDb. Returns title, year, rating, plot, genres, runtime, director, actors, awards, poster and imdburl. |
| Song lyrics | Song | The lyrics of a song. Returns title, artist, lyrics, image and url. |
| iTunes song | Song | A song on iTunes. Returns name, artist, album, genre, length, price, release_date, thumbnail and url. |
| npm package | Package | A package on npm. Returns name, version, description, author, keywords, repository, last_published and downloads_this_year. |
| Country | Country | A country. Returns name, capital, currency, languages, continent, area, phoneCode, tld, neighbours and flag, the URL of its flag. |
| Chemical element | Element | An element of the periodic table. Returns name, symbol, atomic_number, atomic_mass, period, phase, discovered_by, summary and image. |
| Color | Hex Color | A color. Returns hex, name, rgb, brightened, a lighter version, and color_image, the URL of an image of it. |

<NodeInfoExplorer type="action_popcat_search" />
