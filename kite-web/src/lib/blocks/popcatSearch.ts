import { z } from "zod";
import {
  popcatBlock,
  popcatOperation as op,
  popcatPath,
  popcatQuery,
} from "./popcat";

const search = (label: string, description: string) =>
  popcatQuery("popcat_text", "q", label, description);

export const popcatSearch = popcatBlock({
  type: "action_popcat_search",
  id_prefix: "popcat",
  title: "Look up with Popcat",
  description: "Look up the weather, a GitHub user, a movie, a song and more",
  icon: "search",
  setting_label: "Look Up",
  operations: [
    op(
      "/v2/weather",
      "Weather",
      "The weather of a place. Returns a list of matching places, each with location.name, location.degreetype, current.temperature, current.feelslike, current.skytext, current.humidity, current.winddisplay, current.imageUrl and forecast, a list of days with day, low, high, skytextday and precip.",
      search("Place", "The place, like Chicago or Paris, France.")
    ),
    op(
      "/v2/github/{username}",
      "GitHub user",
      "A GitHub user. Returns url, avatar, name, bio, company, location, blog, followers, following, public_repos and created_at.",
      popcatPath("popcat_text", "username", "Username", "The GitHub username.")
    ),
    op(
      "/v2/steam",
      "Steam game",
      "A game on Steam. Returns name, description, price, thumbnail, banner, website, developers and publishers.",
      search("Game", "Name of the game.")
    ),
    op(
      "/v2/imdb",
      "Movie or show",
      "A movie or show on IMDb. Returns title, year, rating, plot, genres, runtime, director, actors, awards, poster and imdburl.",
      search("Title", "Title of the movie or show.")
    ),
    op(
      "/v2/lyrics",
      "Song lyrics",
      "The lyrics of a song. Returns title, artist, lyrics, image and url.",
      popcatQuery(
        "popcat_text",
        "song",
        "Song",
        "Name of the song, optionally with the artist."
      )
    ),
    op(
      "/v2/itunes",
      "iTunes song",
      "A song on iTunes. Returns name, artist, album, genre, length, price, release_date, thumbnail and url.",
      search("Song", "Name of the song.")
    ),
    op(
      "/v2/npm",
      "npm package",
      "A package on npm. Returns name, version, description, author, keywords, repository, last_published and downloads_this_year.",
      search("Package", "Name of the package, like discord.js.")
    ),
    op(
      "/v2/countries/{name}",
      "Country",
      "A country. Returns name, capital, currency, languages, continent, area, phoneCode, tld, neighbours and flag, the URL of its flag.",
      popcatPath("popcat_text", "name", "Country", "Name of the country.")
    ),
    op(
      "/v2/periodic-table",
      "Chemical element",
      "An element of the periodic table. Returns name, symbol, atomic_number, atomic_mass, period, phase, discovered_by, summary and image.",
      popcatQuery(
        "popcat_text",
        "element",
        "Element",
        "Name, symbol or atomic number of the element, like H or Hydrogen."
      )
    ),
    op(
      "/v2/color/{hex}",
      "Color",
      "A color. Returns hex, name, rgb, brightened, a lighter version, and color_image, the URL of an image of it.",
      popcatPath(
        "popcat_text",
        "hex",
        "Hex Color",
        "The color, without #, like ff0000."
      )
    ),
  ],
  result: {
    schema: z
      .unknown()
      .describe(
        "What Popcat returned, with the fields popcat_endpoint lists, like result('id').name. A list of places for the weather."
      ),
  },
});
