import argparse
import logging
import sys
import json
from praatio import textgrid

logger = logging.getLogger("")
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s"
)


def main(argv):
    logger.info("Starting")
    parser = argparse.ArgumentParser(description="Json to praat textgrid")
    parser.add_argument("--input", nargs='?', required=True, help="train data")
    parser.add_argument("--output", nargs='?', required=True, help="ONN model")

    args = parser.parse_args(args=argv)

    logger.info(f"Input JSON data: {args.input}")

    with open(args.input, "r", encoding="utf-8") as f:
        data = json.load(f)

        # [{"timeMillis": 185, "durationMillis": 604, "type": "word", "value": "Pirmasis"},
        #  {"timeMillis": 1068, "durationMillis": 603, "type": "word", "value": "antrasis"},

    sm = data.get("speechMarks",  [])
    logger.info(f"Loaded {len(sm)} speech marks")

    word_entries = []
    end = 0.0
    for m in sm:
        if m.get("type") != "word":
            continue

        start = m["timeMillis"] / 1000.0
        if start < end:
            logger.warn(f"Overlapping words: {start} < {end}, adjusting")
            start = end
        end = (m["timeMillis"] + m["durationMillis"]) / 1000.0
        word_entries.append((start, end, m["value"]))

    end += 3.0  # to avoid issues with praat
    tg = textgrid.Textgrid()
    word_tier = textgrid.IntervalTier(
        name="words",
        entries=word_entries,
        minT=0.0,
        maxT=end    
    )

    tg.addTier(word_tier)

    tg.save(
        args.output,
        format="long_textgrid",
        includeBlankSpaces=True
    )

    logger.info("✅ Done")

if __name__ == "__main__":
    main(sys.argv[1:])