package hclog

import (
	"regexp"
	"testing"
)

func TestExclude(t *testing.T) {
	t.Run("excludes by message", func(t *testing.T) {
		var em ExcludeByMessage
		em.Add("foo")
		em.Add("bar")

		assertTrue(t, em.Exclude(Info, "foo"))
		assertTrue(t, em.Exclude(Info, "bar"))
		assertFalse(t, em.Exclude(Info, "qux"))
		assertFalse(t, em.Exclude(Info, "foo qux"))
		assertFalse(t, em.Exclude(Info, "qux bar"))
	})

	t.Run("excludes by prefix", func(t *testing.T) {
		ebp := ExcludeByPrefix("foo: ")

		assertTrue(t, ebp.Exclude(Info, "foo: rocks"))
		assertFalse(t, ebp.Exclude(Info, "foo"))
		assertFalse(t, ebp.Exclude(Info, "qux foo: bar"))
	})

	t.Run("exclude by regexp", func(t *testing.T) {
		ebr := &ExcludeByRegexp{
			Regexp: regexp.MustCompile("(foo|bar)"),
		}

		assertTrue(t, ebr.Exclude(Info, "foo"))
		assertTrue(t, ebr.Exclude(Info, "bar"))
		assertTrue(t, ebr.Exclude(Info, "foo qux"))
		assertTrue(t, ebr.Exclude(Info, "qux bar"))
		assertFalse(t, ebr.Exclude(Info, "qux"))
	})

	t.Run("excludes many funcs", func(t *testing.T) {
		ef := ExcludeFuncs{
			ExcludeByPrefix("foo: ").Exclude,
			ExcludeByPrefix("bar: ").Exclude,
		}

		assertTrue(t, ef.Exclude(Info, "foo: rocks"))
		assertTrue(t, ef.Exclude(Info, "bar: rocks"))
		assertFalse(t, ef.Exclude(Info, "foo"))
		assertFalse(t, ef.Exclude(Info, "qux foo: bar"))

	})
}
