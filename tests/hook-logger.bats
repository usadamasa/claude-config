#!/usr/bin/env bats
# hooks/lib/hook-logger.sh のテスト

LOGGER_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/hooks/lib/hook-logger.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  export TEST_TMPDIR
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
}

# =============================================================================
# is_dry_run
# =============================================================================

@test "is_dry_run: DRY_RUN=1 で true を返す" {
  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; is_dry_run"

  [ "$status" -eq 0 ]
}

@test "is_dry_run: DRY_RUN=0 で false を返す" {
  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; is_dry_run"

  [ "$status" -ne 0 ]
}

@test "is_dry_run: DRY_RUN 未設定で false を返す" {
  run env -u DRY_RUN HOOK_NAME=test bash -c "source '$LOGGER_PATH'; is_dry_run"

  [ "$status" -ne 0 ]
}

# =============================================================================
# _log_prefix
# =============================================================================

@test "_log_prefix: 通常モードで [HOOK_NAME] を返す" {
  run env DRY_RUN=0 HOOK_NAME=my-hook bash -c "source '$LOGGER_PATH'; _log_prefix"

  [ "$status" -eq 0 ]
  [ "$output" = "[my-hook]" ]
}

@test "_log_prefix: dry-run モードで [DRY-RUN] を返す" {
  run env DRY_RUN=1 HOOK_NAME=my-hook bash -c "source '$LOGGER_PATH'; _log_prefix"

  [ "$status" -eq 0 ]
  [ "$output" = "[DRY-RUN]" ]
}

# =============================================================================
# log_info
# =============================================================================

@test "log_info: メッセージが stderr に出力される" {
  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; log_info 'hello world' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[test]"* ]]
  [[ "$output" == *"INFO"* ]]
  [[ "$output" == *"hello world"* ]]
}

# =============================================================================
# log_skip
# =============================================================================

@test "log_skip: dry-run 時に SKIP メッセージが stderr に出力される" {
  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; log_skip 'no parent memory' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"SKIP"* ]]
  [[ "$output" == *"no parent memory"* ]]
}

@test "log_skip: 通常モードでも出力される" {
  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; log_skip 'no parent memory' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[test]"* ]]
  [[ "$output" == *"SKIP"* ]]
}

# =============================================================================
# logged_mkdir (通常モード)
# =============================================================================

@test "logged_mkdir: 通常モードでディレクトリが作成される" {
  local TARGET="$TEST_TMPDIR/new-dir/sub"

  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_mkdir '$TARGET'"

  [ "$status" -eq 0 ]
  [ -d "$TARGET" ]
}

@test "logged_mkdir: 通常モードで MKDIR ログが stderr に出力される" {
  local TARGET="$TEST_TMPDIR/new-dir2"

  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_mkdir '$TARGET' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[test]"* ]]
  [[ "$output" == *"MKDIR"* ]]
  [[ "$output" == *"$TARGET"* ]]
}

# =============================================================================
# logged_mkdir (dry-run モード)
# =============================================================================

@test "logged_mkdir: dry-run でディレクトリが作成されない" {
  local TARGET="$TEST_TMPDIR/should-not-exist"

  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_mkdir '$TARGET'"

  [ "$status" -eq 0 ]
  [ ! -d "$TARGET" ]
}

@test "logged_mkdir: dry-run で [DRY-RUN] MKDIR ログが出る" {
  local TARGET="$TEST_TMPDIR/dry-dir"

  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_mkdir '$TARGET' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"MKDIR"* ]]
}

# =============================================================================
# logged_cp (通常モード)
# =============================================================================

@test "logged_cp: 通常モードでファイルがコピーされる" {
  echo "content" > "$TEST_TMPDIR/src.txt"

  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cp '$TEST_TMPDIR/src.txt' '$TEST_TMPDIR/dst.txt'"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/dst.txt" ]
  [ "$(cat "$TEST_TMPDIR/dst.txt")" = "content" ]
}

@test "logged_cp: 通常モードで CP ログが stderr に出力される" {
  echo "content" > "$TEST_TMPDIR/src2.txt"

  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cp '$TEST_TMPDIR/src2.txt' '$TEST_TMPDIR/dst2.txt' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[test]"* ]]
  [[ "$output" == *"CP"* ]]
  [[ "$output" == *"src2.txt"* ]]
  [[ "$output" == *"dst2.txt"* ]]
}

# =============================================================================
# logged_cp (dry-run モード)
# =============================================================================

@test "logged_cp: dry-run でファイルがコピーされない" {
  echo "content" > "$TEST_TMPDIR/src3.txt"

  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cp '$TEST_TMPDIR/src3.txt' '$TEST_TMPDIR/dst3.txt'"

  [ "$status" -eq 0 ]
  [ ! -f "$TEST_TMPDIR/dst3.txt" ]
}

@test "logged_cp: dry-run で [DRY-RUN] CP ログが出る" {
  echo "content" > "$TEST_TMPDIR/src4.txt"

  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cp '$TEST_TMPDIR/src4.txt' '$TEST_TMPDIR/dst4.txt' 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"CP"* ]]
}

# =============================================================================
# logged_cmd (通常モード)
# =============================================================================

@test "logged_cmd: 通常モードでコマンドが実行される" {
  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cmd touch '$TEST_TMPDIR/created.txt'"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/created.txt" ]
}

@test "logged_cmd: 通常モードで CMD ログが stderr に出力される" {
  run env DRY_RUN=0 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cmd echo hello 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[test]"* ]]
  [[ "$output" == *"CMD"* ]]
  [[ "$output" == *"echo hello"* ]]
}

# =============================================================================
# logged_cmd (dry-run モード)
# =============================================================================

@test "logged_cmd: dry-run でコマンドが実行されない" {
  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cmd touch '$TEST_TMPDIR/should-not-exist.txt'"

  [ "$status" -eq 0 ]
  [ ! -f "$TEST_TMPDIR/should-not-exist.txt" ]
}

@test "logged_cmd: dry-run で [DRY-RUN] CMD ログが出る" {
  run env DRY_RUN=1 HOOK_NAME=test bash -c "source '$LOGGER_PATH'; logged_cmd git wt --nocd my-branch 2>&1"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"CMD"* ]]
  [[ "$output" == *"git wt --nocd my-branch"* ]]
}
