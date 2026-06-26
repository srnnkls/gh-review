package cmd

import "testing"

func TestResolveThreadID(t *testing.T) {
	t.Run("thread flag short-circuits without a client", func(t *testing.T) {
		got, err := resolveThreadID(nil, nil, "  PRRT_x  ", "")
		if err != nil {
			t.Fatalf("resolveThreadID() unexpected error: %v", err)
		}
		if got != "PRRT_x" {
			t.Errorf("threadID = %q, want %q", got, "PRRT_x")
		}
	})

	t.Run("thread flag wins over comment flag", func(t *testing.T) {
		got, err := resolveThreadID(nil, nil, "PRRT_x", "PRRC_y")
		if err != nil {
			t.Fatalf("resolveThreadID() unexpected error: %v", err)
		}
		if got != "PRRT_x" {
			t.Errorf("threadID = %q, want %q", got, "PRRT_x")
		}
	})

	t.Run("neither flag is an error", func(t *testing.T) {
		_, err := resolveThreadID(nil, nil, "", "  ")
		if err == nil {
			t.Error("resolveThreadID() expected error when neither --thread nor --comment is set")
		}
	})
}

func TestReplyResolveCmdFlags(t *testing.T) {
	if replyCmd.Flags().Lookup("body") == nil {
		t.Error("reply: body flag not registered")
	}
	if replyCmd.Flags().Lookup("comment").Shorthand != "c" {
		t.Error("reply: comment flag shorthand should be c")
	}
	if resolveCmd.Flags().Lookup("comment").Shorthand != "c" {
		t.Error("resolve: comment flag shorthand should be c")
	}
	if resolveCmd.Flags().Lookup("thread") == nil {
		t.Error("resolve: thread flag not registered")
	}
}
