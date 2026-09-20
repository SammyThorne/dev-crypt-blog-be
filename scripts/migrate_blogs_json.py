#!/usr/bin/env python3
"""Migrate a legacy blogs.json file into the `posts` table.

Usage:
    python3 migrate_blogs_json.py blogs.json [--author-uuid UUID] [--tz ZONE] [--dry-run]

Expects a JSON array of objects shaped like:
    {
      "title": "...",
      "blurb": "...",
      "dateTime": "2026/05/24 8:25 AM",   (optional, defaults to now)
      "content": "..."                    (optional, defaults to "")
    }

The migration is idempotent: a post is skipped if one with the same title and
date_time already exists, so the script can safely be re-run. With --dry-run
nothing is written; the transaction is rolled back after reporting.

Connects using the same env vars (and defaults) as the Go backend:
DB_HOST, DB_USER, DB_PASSWORD, DB_NAME.
"""
import argparse
import json
import os
import sys
from datetime import datetime
from zoneinfo import ZoneInfo

import psycopg2

DATE_FORMAT = "%Y/%m/%d %I:%M %p"


def parse_args():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("json_file", help="Path to the legacy blogs.json file")
    parser.add_argument(
        "--author-uuid",
        default=None,
        help="Firebase UID to attribute the migrated posts to (optional)",
    )
    parser.add_argument(
        "--tz",
        default="UTC",
        help="IANA timezone the dateTime values are written in (default: UTC)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Report what would be inserted without writing to the database",
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

    try:
        tz = ZoneInfo(args.tz)
    except Exception:
        sys.exit(f"Unknown timezone '{args.tz}'")

    db_password = os.environ.get("DB_PASSWORD")
    if not db_password:
        sys.exit("DB_PASSWORD environment variable is required")

    conn = psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        user=os.environ.get("DB_USER", "postgres"),
        password=db_password,
        dbname=os.environ.get("DB_NAME", "test_db"),
    )

    inserted = duplicates = invalid = 0
    try:
        with conn.cursor() as cur:
            for i, post in enumerate(posts):
                title = post.get("title")
                blurb = post.get("blurb")
                content = post.get("content") or ""
                raw_date = post.get("dateTime")

                if not title or not blurb:
                    print(f"Skipping entry {i}: missing title or blurb", file=sys.stderr)
                    invalid += 1
                    continue

                if raw_date:
                    try:
                        date_time = datetime.strptime(raw_date, DATE_FORMAT).replace(tzinfo=tz)
                    except ValueError:
                        print(
                            f"Skipping entry {i} ('{title}'): unparseable dateTime "
                            f"'{raw_date}', expected format like '2026/05/24 8:25 AM'",
                            file=sys.stderr,
                        )
                        invalid += 1
                        continue
                else:
                    # No date in the source, so fall back to "now". Note this
                    # defeats the duplicate check on re-runs for such entries.
                    date_time = datetime.now(tz)

                cur.execute(
                    """
                    INSERT INTO posts (title, blurb, content, date_time, author_uuid)
                    SELECT %s, %s, %s, %s, %s
                    WHERE NOT EXISTS (
                        SELECT 1 FROM posts WHERE title = %s AND date_time = %s
                    )
                    """,
                    (title, blurb, content, date_time, args.author_uuid, title, date_time),
                )
                if cur.rowcount:
                    inserted += 1
                    print(f"{'Would insert' if args.dry_run else 'Inserted'}: {title}")
                else:
                    duplicates += 1
                    print(f"Already exists, skipped: {title}")

        if args.dry_run:
            conn.rollback()
        else:
            conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()

    verb = "Would insert" if args.dry_run else "Inserted"
    print(
        f"{verb} {inserted} of {len(posts)} post(s); "
        f"{duplicates} already present, {invalid} invalid."
    )


if __name__ == "__main__":
    main()
