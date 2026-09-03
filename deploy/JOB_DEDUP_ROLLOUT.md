# Job deduplication rollout

This release makes a non-seed account job unique by `(user_id, lower(info_hash))`
while it is live. Failed and evicted history, private seed jobs, and the internal
`system`/`prewarm` users are intentionally outside the constraint.

`shared/jobs/schema.sql` performs the migration when the first updated service
opens the jobs database. It takes a `SHARE ROW EXCLUSIVE` lock so reads continue
but job inserts/updates pause until cleanup and index creation commit.

## Before deployment

Take the normal database snapshot, then measure the rows that will be
consolidated:

```sql
SELECT sum(count - 1) AS rows_to_archive
FROM (
    SELECT count(*)
    FROM jobs
    WHERE user_id NOT IN ('', 'system', 'prewarm')
      AND info_hash <> ''
      AND seed = false
      AND status NOT IN ('failed', 'evicted')
    GROUP BY user_id, lower(info_hash)
    HAVING count(*) > 1
) duplicates;
```

For a very large result, deploy during a low-write window. Do not run multiple
old-version API replicas for longer than necessary after the first updated
replica starts: the new index prevents their duplicate writes, but those old
replicas may return an insert error until replaced.

## Deployment order

1. Drain or stop new job submissions if the jobs table is large.
2. Start one updated API/service instance and wait for database initialization.
3. Verify the index and duplicate count below.
4. Roll the remaining API, Stremio, worker, and companion services.
5. Re-enable submissions and test one movie plus two episodes from one season
   pack.

The separate `torrin-addon` does not require a coordinated release; it already
sends `sid=tt...:season:episode` to the StremThru endpoints.

## Verification

```sql
SELECT indexdef
FROM pg_indexes
WHERE indexname = 'idx_jobs_one_live_user_hash';

SELECT user_id, lower(info_hash), count(*)
FROM jobs
WHERE user_id NOT IN ('', 'system', 'prewarm')
  AND info_hash <> ''
  AND seed = false
  AND status NOT IN ('failed', 'evicted')
GROUP BY user_id, lower(info_hash)
HAVING count(*) > 1;

SELECT count(*) AS archived_duplicates
FROM job_dedup_archive;
```

The duplicate query must return no rows. Replaying the same cached movie or a
different episode from the same pack must keep the same job ID.

## Rollback

Application rollback does not require restoring archived duplicates. If a full
data rollback is explicitly required, first stop job writes and drop the index,
then restore archived rows:

```sql
DROP INDEX IF EXISTS idx_jobs_one_live_user_hash;

INSERT INTO jobs
SELECT (jsonb_populate_record(NULL::jobs, row_data)).*
FROM job_dedup_archive
ON CONFLICT (id) DO NOTHING;
```

Keep `job_dedup_archive` until the release has been stable through the normal
rollback window.
