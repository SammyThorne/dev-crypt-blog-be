#!/usr/bin/env python3
"""Bulk-insert blog posts from a JSON file into the `posts` table.

Usage:
    python3 import_posts.py posts.json [--author-uuid UUID]

Expects a JSON array of objects shaped like:
    {
      "title": "...",
      "blurb": "...",
      "dateTime": "2026/05/24 8:25 AM",
      "content": "..."   (optional, defaults to "")
    }

Connects using the same env vars (and defaults) as the Go backend:
DB_HOST, DB_USER, DB_PASSWORD, DB_NAME.
"""
import argparse
import json
import os
import sys
from datetime import datetime

import psycopg2

DATE_FORMAT = "%Y/%m/%d %I:%M %p"


def parse_args():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("json_file", help="Path to the JSON file of posts")
    parser.add_argument(
        "--author-uuid",
        default=None,
        help="Firebase UID to attribute the imported posts to (optional)",
    )
    return parser.parse_args()


def load_posts(path):
    with open(path, "r", encoding="utf-8") as f:
        posts = json.load(f)
    if not isinstance(posts, list):
        raise ValueError("Expected a JSON array of post objects")
    return posts


def main():
    args = parse_args()
    posts = load_posts(args.json_file)

    db_password = os.environ.get("DB_PASSWORD")
    if not db_password:
        sys.exit("DB_PASSWORD environment variable is required")

    conn = psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        user=os.environ.get("DB_USER", "postgres"),
        password=db_password,
        dbname=os.environ.get("DB_NAME", "test_db"),
    )

    inserted = 0
    try:
        with conn:
            with conn.cursor() as cur:
                for i, post in enumerate(posts):
                    title = post.get("title")
                    blurb = post.get("blurb")
                    content = post.get("content", "")
                    raw_date = post.get("dateTime")

                    if not title or not blurb:
                        print(f"Skipping entry {i}: missing title or blurb", file=sys.stderr)
                        continue

                    date_time = None
                    if raw_date:
                        try:
                            date_time = datetime.strptime(raw_date, DATE_FORMAT)
                        except ValueError:
                            print(
                                f"Skipping entry {i} ('{title}'): unparseable dateTime "
                                f"'{raw_date}', expected format like '2026/05/24 8:25 AM'",
                                file=sys.stderr,
                            )
                            continue

                    cur.execute(
                        """
                        INSERT INTO posts (title, blurb, content, date_time, author_uuid)
                        VALUES (%s, %s, %s, COALESCE(%s, CURRENT_TIMESTAMP), %s)
                        """,
                        (title, blurb, content, date_time, args.author_uuid),
                    )
                    inserted += 1
    finally:
        conn.close()

    print(f"Inserted {inserted} of {len(posts)} post(s).")


if __name__ == "__main__":
    main()
