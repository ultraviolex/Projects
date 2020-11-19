package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	"reflect"
	_ "strconv"
	_ "strings"
	"testing"

	"github.com/cs161-staff/userlib"
	"github.com/google/uuid"
)

// *************************** HELPERS ****************************** //

// Wipes storage and sets debug status to true (call before each test)
func clear() {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	// Turn on debugging for all tests
	userlib.SetDebugStatus(true)
}

// Gets a string full of random bytes of specified length (in chars)
func randomString(numbytes int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"!@#$%%^&*()_+}{\":?><|~/.,';\\][=-]}"
	var numchars = len(chars)

	var bytes = userlib.RandomBytes(numbytes)
	var ret []byte = make([]byte, numbytes)

	for i, val := range bytes {
		ret[i] = chars[int(val)%numchars]
	}
	return string(ret)
}

// **************************** USER TESTS ************************** //

func TestUser1(t *testing.T) {
	clear()

	_, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
}

func TestUser2(t *testing.T) {
	clear()

	_, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize first alice user", err)
		return
	}
	_, err = InitUser("alice", "rabuf")
	if err == nil {
		t.Error("Failed to check if username already taken", err)
		return
	}
	_, err = InitUser("Alice", "foo")
	if err != nil {
		t.Error("Did not allow creation of users with usernames differing only in case", err)
		return
	}
}

func TestUser3(t *testing.T) {
	clear()

	const count = 5 // number of times to repeat random username and password test

	_, err := InitUser("alice", "")
	if err != nil {
		t.Error("Failed to initialize user with blank password", err)
		return
	}
	_, err = InitUser("", "baz")
	if err != nil {
		t.Error("Failed to initialize user with blank username", err)
		return
	}

	for i := 0; i < count; i++ {
		_, err = InitUser(randomString(4096), randomString(4096))
		if err != nil {
			t.Error("Failed to initialize user with extremely long, random username and password")
			return
		}
	}
}

func TestUser4(t *testing.T) {
	clear()

	u1, _ := InitUser("alice", "foo")
	u2, _ := InitUser("Alice", "foo")
	if reflect.DeepEqual(u1, u2) {
		t.Error("Key generation is deterministic from password")
		return
	}

	clear()
	u3, _ := InitUser("alice", "foo")
	if reflect.DeepEqual(u1, u3) {
		t.Error("Key generation is deterministic from password and username")
		return
	}
}

func TestUser5(t *testing.T) {
	clear()
	u1, _ := InitUser("alice", "foo")
	u2, _ := InitUser("alic", "efoo")
	u3, _ := InitUser("lice", "fooa")
	if reflect.DeepEqual(u1, u2) || reflect.DeepEqual(u1, u3) {
		t.Error("Username and password are concatenated directly")
		return
	}

	clear()
	u4, _ := InitUser("bob-", "bar")
	u5, _ := InitUser("bob", "-bar")
	if reflect.DeepEqual(u4, u5) {
		t.Error("Username and password are concatenated with a - in between")
		return
	}

	clear()
	u6, _ := InitUser("-bob", "bar")
	u7, _ := InitUser("bob", "bar-")
	if reflect.DeepEqual(u6, u7) {
		t.Error("Username and password are concatenated with a - in between")
		return
	}
}

func TestUser6(t *testing.T) {
	clear()

	u1, _ := InitUser("alice", "foo")
	u2, err := GetUser("alice", "foo")
	if err != nil {
		t.Error("Failed to get alice from database")
		return
	}
	if !reflect.DeepEqual(u1, u2) {
		t.Error("User returned from Init and Get are different")
		return
	}
	_, err = GetUser("alice", "bar")
	if err == nil {
		t.Error("GetUser did not error on wrong password")
		return
	}
	_, err = GetUser("bob", "fubar")
	if err == nil {
		t.Error("GetUser did not error on nonexistent user")
		return
	}
}

func TestUser7(t *testing.T) {
	clear()
	username := randomString(1024)
	password := randomString(1024)
	u1, _ := InitUser(username, password)
	u2, _ := GetUser(username, password)

	if !reflect.DeepEqual(u1, u2) {
		t.Error("User returned from Init and Get are different with random long string")
		return
	}

	clear()
	u1, err := InitUser("", "")
	if err != nil {
		t.Error("Unable to initialize user with blank username and password")
		return
	}
	u2, _ = GetUser("", "")

	if !reflect.DeepEqual(u1, u2) {
		t.Error("User returned from Init and Get are different with blank username and password")
		return
	}
}

func TestUser8(t *testing.T) {
	clear()

	var datastore map[uuid.UUID][]byte

	u1, _ := InitUser("alice", "foo")
	datastore = userlib.DatastoreGetMap()

	userlib.DatastoreClear()
	for k, v := range datastore {
		userlib.DatastoreSet(k, v)
	}
	u2, err := GetUser("alice", "foo")
	if err != nil {
		t.Error("GetUser failed after resetting client; client is not stateless")
		return
	}
	if !reflect.DeepEqual(u1, u2) {
		t.Error("User returned from Init and Get are different when client is reset in between")
		return
	}
}

// ***************************** FILE TESTS ************************ //

func TestFile1(t *testing.T) {
	clear()

	u, _ := InitUser("alice", "fubar")

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestFile2(t *testing.T) {
	clear()

	u, _ := InitUser("alice", "fubar")
	v, err2 := u.LoadFile("this file does not exist")
	if err2 == nil {
		t.Error("Downloaded a nonexistent file", err2)
		return
	}
	if v != nil {
		t.Error("LoadFile returned bogus data for nonexistent file")
		return
	}
}

func TestFile3(t *testing.T) {
	clear()

	a, _ := InitUser("alice", "foo")
	b, _ := InitUser("bob", "bar")
	va := []byte("Hello, this is Alice")
	vb := []byte("Hello, this is Bob")

	a.StoreFile("f", va)
	b.StoreFile("f", vb)

	reta, err := a.LoadFile("f")
	if err != nil {
		t.Error("Failed to download Alice's file")
		return
	}
	if !reflect.DeepEqual(reta, va) {
		t.Error("Downloaded file is not the same as original for Alice", reta, va)
		return
	}

	retb, err := b.LoadFile("f")
	if err != nil {
		t.Error("Failed to download Bob's file")
		return
	}
	if !reflect.DeepEqual(retb, vb) {
		t.Error("Downloaded file is not the same as original for Bob", retb, vb)
		return
	}
}

func TestFile4(t *testing.T) {
	clear()

	a, _ := InitUser("alice", "foo")
	v1 := []byte("This is version 1")
	v2 := []byte("This is version 2")
	v3 := []byte("This is version 3")

	a.StoreFile("f", v1)
	a.StoreFile("f", v2)
	a.StoreFile("f", v3)

	ret, err := a.LoadFile("f")
	if err != nil {
		t.Error("Failed to downloaded Alice's file \"f\"")
		return
	}
	if !reflect.DeepEqual(ret, v3) {
		t.Error("Downloaded file does not contain last-stored version of file", ret, v3)
		return
	}
}

func TestFile5(t *testing.T) {
	clear()

	a1, _ := InitUser("alice", "foo")
	a2, _ := GetUser("alice", "foo")
	v1 := []byte("This is version 1")
	v2 := []byte("This is version 2")
	v3 := []byte("This is version 3")

	a1.StoreFile("f", v1)
	ret1, _ := a1.LoadFile("f")
	ret2, err := a2.LoadFile("f")
	if err != nil {
		t.Error("Other instance of user could not download uploaded file")
		return
	}

	a2.StoreFile("f", v2)
	ret2, _ = a2.LoadFile("f")
	ret1, err = a1.LoadFile("f")
	if err != nil {
		t.Error("Original instance of user could not download updated file")
		return
	}
	if !reflect.DeepEqual(ret1, ret2) || !reflect.DeepEqual(ret1, v2) {
		t.Error("Original instance of user did not download updated version of file")
		return
	}

	a1.StoreFile("f", v3)
	ret1, _ = a1.LoadFile("f")
	ret2, err = a2.LoadFile("f")
	if err != nil {
		t.Error("Other instance of user could not download updated file")
		return
	}
	if !reflect.DeepEqual(ret1, ret2) || !reflect.DeepEqual(ret2, v3) {
		t.Error("Other instance of user did not download updated version of file")
		return
	}
}

func TestFile6(t *testing.T) {
	clear()
	const count = 16 // number of times to repeat this test
	var filename string
	var data []byte

	a, _ := InitUser("alice", "foo")

	for i := 0; i < count; i++ {
		filename = randomString(4096)
		data = userlib.RandomBytes(8192)
		a.StoreFile(filename, data)
		ret, err := a.LoadFile(filename)
		if err != nil {
			t.Error("Failed to load file with extremely long name")
			return
		}
		if !reflect.DeepEqual(ret, data) {
			t.Error("Failed to load correct contents of file with extremely long name")
			return
		}
	}
}

func TestFile7(t *testing.T) {
	clear()

	a, _ := InitUser("alice", "foo")
	a.StoreFile("", nil)
	ret, err := a.LoadFile("")
	if err != nil {
		t.Error("Failed to load file with blank name")
		return
	}
	if len(ret) != 0 {
		t.Error("Failed to load 0 bytes for file with no contents")
		return
	}

	a.StoreFile("_", make([]byte, 0))
	ret, err = a.LoadFile("_")
	if err != nil {
		t.Error("Failed to load file with single char name")
		return
	}
	if len(ret) != 0 {
		t.Error("Failed to load 0 bytes for file with no contents")
		return
	}
}

func TestFile8(t *testing.T) {
	clear()

	a, _ := InitUser("alice", "foo")
	v1 := []byte("This is line 1")
	v2 := []byte("This is line 2")
	v3 := []byte("This is line 3")
	concat := append(v1, v2...)
	concat = append(concat, v3...)

	err := a.AppendFile("f", v1)
	if err == nil {
		t.Error("Allowed appending to a nonexistent file")
		return
	}

	a.StoreFile("f", v1)
	err = a.AppendFile("f", v2)
	if err != nil {
		t.Error("Failed to append line 2 to file")
		return
	}
	err = a.AppendFile("f", v3)
	if err != nil {
		t.Error("Failed to append line 3 to file")
		return
	}

	ret, err := a.LoadFile("f")
	if err != nil {
		t.Error("Failed to download file that has been appended to")
		return
	}
	if !reflect.DeepEqual(ret, concat) {
		t.Error("Failed to load correct contents for file that has been appended to")
		return
	}
}

func TestFile9(t *testing.T) {
	clear()

	a1, _ := InitUser("alice", "foo")
	a2, _ := GetUser("alice", "foo")
	v1 := []byte("This is line 1")
	v2 := []byte("This is line 2")
	v3 := []byte("This is line 3")
	concat := append(v1, v2...)
	concat = append(concat, v3...)

	a1.StoreFile("f", v1)
	a1.AppendFile("f", v2)
	ret, err := a2.LoadFile("f")
	if err != nil {
		t.Error("Other user failed to load appended file")
		return
	}
	if !reflect.DeepEqual(ret, append(v1, v2...)) {
		t.Error("Other user failed to load correct contents of appended file")
		return
	}

	a2.StoreFile("f", append(v1, v2...))
	ret, err = a1.LoadFile("f")
	if err != nil {
		t.Error("Original user failed to load newly stored file")
		return
	}
	if !reflect.DeepEqual(ret, append(v1, v2...)) {
		t.Error("Original user failed to load correct contents of newly stored file")
		return
	}

	err = a1.AppendFile("f", v3)
	if err != nil {
		t.Error("Original user failed to append to file a second time")
		return
	}
	ret, err = a2.LoadFile("f")
	if err != nil {
		t.Error("Other user failed to load appended file #2")
		return
	}
	if !reflect.DeepEqual(ret, concat) {
		t.Error("Other user failed to load correct contents of appended file #2")
		return
	}
	ret, err = a1.LoadFile("f")
	if err != nil {
		t.Error("Original user failed to load appended file #2")
		return
	}
	if !reflect.DeepEqual(ret, concat) {
		t.Error("Original user failed to load correct contents of appended file #2")
		return
	}
}

// **************************** SHARE TESTS ************************ //

func TestShare(t *testing.T) {
	clear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	var v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
}

// copy of TestShare
func TestShare1(t *testing.T) {
	clear()
	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice")
		return
	}
	bob, err := InitUser("bob", "foobar")
	if err != nil {
		t.Error("Failed to initialize bob")
		return
	}

	v := []byte("Testing")
	alice.StoreFile("file1", v)
	v, err = alice.LoadFile("file1")
	if err != nil {
		t.Error("alice's file1 stored incorrectly")
		return
	}
	var magic_string string
	var v2 []byte
	magic_string, err = alice.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Alice failed to share file1(v) with bob")
		return
	}

	err = bob.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	v2, err = bob.LoadFile("file2")
	if err != nil {
		t.Error("Bob failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
}

// test if revoked user still has access to file
func TestShare2(t *testing.T) {
	clear()
	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice")
		return
	}
	bob, err := InitUser("bob", "foobar")
	if err != nil {
		t.Error("Failed to initialize bob")
		return
	}

	v := []byte("Testing")
	alice.StoreFile("file1", v)
	v, err = alice.LoadFile("file1")
	if err != nil {
		t.Error("alice's file1 stored incorrectly")
		return
	}
	var magic_string string
	var v2 []byte
	magic_string, err = alice.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Alice failed to share file1(v) with bob")
		return
	}

	err = bob.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	v2, err = bob.LoadFile("file2")
	if err != nil {
		t.Error("Bob failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	err = alice.RevokeFile("file1", "bob")
	if err != nil {
		t.Error("Alice failed to revoke file", err)
		return
	}

	v2, err = bob.LoadFile("file2")
	if err == nil {
		t.Error("Bob can still access file")
		return
	}

	v2, err = alice.LoadFile("file1")
	if err != nil {
		t.Error("Alice cannot load the file")
		return
	}
}

// test if user can share to a nonexistent user
func TestShare3(t *testing.T) {
	clear()
	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice")
		return
	}
	v := []byte("Testing")
	alice.StoreFile("file1", v)
	v, err = alice.LoadFile("file1")
	if err != nil {
		t.Error("alice's file1 stored incorrectly")
		return
	}
	_, err = alice.ShareFile("file1", "bob")
	if err == nil {
		t.Error("Alice shared file1(v) with a nonexistent user")
		return
	}
}

func TestShare4(t *testing.T) {
	clear()
	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice")
		return
	}
	bob, err := InitUser("bob", "foobar")
	if err != nil {
		t.Error("Failed to initialize bob")
		return
	}
	mallory, err := InitUser("mallory", "foodbar")
	if err != nil {
		t.Error("Failed to initialize mallory")
		return
	}

	v := []byte("Testing")
	alice.StoreFile("file1", v)
	v, err = alice.LoadFile("file1")
	if err != nil {
		t.Error("alice's file1 stored incorrectly")
		return
	}
	var magic_string string
	magic_string, err = alice.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Alice failed to share file1(v) with bob")
		return
	}

	err = mallory.ReceiveFile("file2", "alice", magic_string)
	if err == nil {
		t.Error("Mallory accessed using magic_string", err)
		return
	}

	err = bob.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Bob failed to get share message", err)
		return
	}
}

func TestShare5(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	c, err := InitUser("c", "cc")
	if err != nil {
		t.Error("Failed to initialize c")
		return
	}
	d, err := InitUser("d", "dd")
	if err != nil {
		t.Error("Failed to initialize d")
		return
	}
	e, err := InitUser("e", "ee")
	if err != nil {
		t.Error("Failed to initialize e")
		return
	}
	f, err := InitUser("f", "ff")
	if err != nil {
		t.Error("Failed to initialize f")
		return
	}
	va := []byte("Testing")
	a.StoreFile("file1", va)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	var magic_string_b string
	var magic_string_c string
	var magic_string_d string
	var magic_string_e string
	var magic_string_f string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	magic_string_c, err = a.ShareFile("file1", "c")
	if err != nil {
		t.Error("a failed to share file1 with c")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b failed to recieve file1 from a")
		return
	}
	err = c.ReceiveFile("file1", "a", magic_string_c)
	if err != nil {
		t.Error("c failed to recieve file1 from a")
		return
	}
	magic_string_d, err = b.ShareFile("file1", "d")
	if err != nil {
		t.Error("b failed to share file with d")
		return
	}
	magic_string_e, err = b.ShareFile("file1", "e")
	if err != nil {
		t.Error("b failed to share file with e")
		return
	}
	err = d.ReceiveFile("file1", "b", magic_string_d)
	if err != nil {
		t.Error("d failed to recieve file1 from b")
		return
	}
	err = e.ReceiveFile("file1", "b", magic_string_e)
	if err != nil {
		t.Error("e failed to recieve file1 from b")
		return
	}
	magic_string_f, err = e.ShareFile("file1", "f")
	if err != nil {
		t.Error("e failed to share file with f")
		return
	}
	err = f.ReceiveFile("file1", "e", magic_string_f)
	if err != nil {
		t.Error("f failed to recieve file1 from e")
		return
	}
	vb, err := b.LoadFile("file1")
	if err != nil {
		t.Error("b failed to load file1")
		return
	}
	vc, err := c.LoadFile("file1")
	if err != nil {
		t.Error("c failed to load file1")
		return
	}
	vd, err := d.LoadFile("file1")
	if err != nil {
		t.Error("d failed to load file1")
		return
	}
	ve, err := e.LoadFile("file1")
	if err != nil {
		t.Error("e failed to load file1")
		return
	}
	vf, err := f.LoadFile("file1")
	if err != nil {
		t.Error("f failed to load file1")
		return
	}
	if !reflect.DeepEqual(va, vb) {
		t.Error("Shared file is not the same (va, vb)", va, vb)
		return
	}
	if !reflect.DeepEqual(va, vc) {
		t.Error("Shared file is not the same (va, vc)", va, vc)
		return
	}
	if !reflect.DeepEqual(va, vd) {
		t.Error("Shared file is not the same (va, vd)", va, vd)
		return
	}
	if !reflect.DeepEqual(va, ve) {
		t.Error("Shared file is not the same (va, ve)", va, ve)
		return
	}
	if !reflect.DeepEqual(va, vf) {
		t.Error("Shared file is not the same (va, vf)", va, vf)
		return
	}
	err = a.RevokeFile("file1", "b")
	if err != nil {
		t.Error("a cannot revoke file")
		return
	}
	_, err = b.LoadFile("file1")
	if err == nil {
		t.Error("b can still load file")
		return
	}
	_, err = d.LoadFile("file1")
	if err == nil {
		t.Error("d can still load file")
		return
	}
	_, err = e.LoadFile("file1")
	if err == nil {
		t.Error("e can still load file")
		return
	}
	_, err = f.LoadFile("file1")
	if err == nil {
		t.Error("f can still load file")
		return
	}
	vc, err = c.LoadFile("file1")
	if err != nil {
		t.Error("c cannot load file")
		return
	}
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a cannot load file")
		return
	}
	err = a.RevokeFile("file1", "c")
	if err != nil {
		t.Error("a failed to revoke c")
		return
	}
	_, err = c.LoadFile("file1")
	if err == nil {
		t.Error("c can still load file")
		return
	}
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a cannot load file")
		return
	}
}

func TestShare6(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	va := []byte("Testing")
	a.StoreFile("file1", va)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	var magic_string_b string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b failed to recieve file1 from a")
		return
	}
	err = b.RevokeFile("file1", "a")
	if err == nil {
		t.Error("b can revoke file owned by a")
		return
	}

}

func TestShare7(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	va := []byte("a file1")
	a.StoreFile("file1", va)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	vb := []byte("b file1")
	b.StoreFile("file1", vb)
	var magic_string_b string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err == nil {
		t.Error("b recieved file with identical name from a")
		return
	}

}

func TestShare8(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	v1 := []byte("line1")
	a.StoreFile("file1", v1)
	v1, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	var magic_string_b string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b failed to recieve file from a")
		return
	}
	err = a.RevokeFile("file1", "b")
	if err != nil {
		t.Error("a failed to revoke file from b")
		return
	}
	v2 := []byte("line2")
	err = a.AppendFile("file1", v2)
	if err != nil {
		t.Error("a failed to append to file after revocation")
		return
	}
	err = b.ReceiveFile("file2", "a", magic_string_b)
	vb, err := b.LoadFile("file2")
	if err == nil {
		t.Error("bob can still access revoked file using maging string")
	}
	if reflect.DeepEqual(append(v1, v2...), vb) {
		t.Error("Bob can access revoked file")
		return
	}
}

func TestShare9(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	v1 := []byte("line1")
	a.StoreFile("file1", v1)
	v1, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	var magic_string_b string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b failed to recieve file from a")
		return
	}
	v2 := []byte("line2")
	err = a.AppendFile("file1", v2)
	if err != nil {
		t.Error("a failed to append to file")
		return
	}
	vb, err := b.LoadFile("file1")
	if err != nil {
		t.Error("b cannot access file after append")
		return
	}
	if !reflect.DeepEqual(append(v1, v2...), vb) {
		t.Error("b's file not updated can access revoked file")
		return
	}
}

func TestShare10(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	c, err := InitUser("c", "cc")
	if err != nil {
		t.Error("Failed to initialize c")
		return
	}
	d, err := InitUser("d", "dd")
	if err != nil {
		t.Error("Failed to initialize d")
		return
	}
	e, err := InitUser("e", "ee")
	if err != nil {
		t.Error("Failed to initialize e")
		return
	}
	f, err := InitUser("f", "ff")
	if err != nil {
		t.Error("Failed to initialize f")
		return
	}
	va := []byte("Testing")
	a.StoreFile("file1", va)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	var magic_string_b string
	var magic_string_c string
	var magic_string_d string
	var magic_string_e string
	var magic_string_f string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	magic_string_c, err = a.ShareFile("file1", "c")
	if err != nil {
		t.Error("a failed to share file1 with c")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b failed to recieve file1 from a")
		return
	}
	err = c.ReceiveFile("file1", "a", magic_string_c)
	if err != nil {
		t.Error("c failed to recieve file1 from a")
		return
	}
	magic_string_d, err = b.ShareFile("file1", "d")
	if err != nil {
		t.Error("b failed to share file with d")
		return
	}
	magic_string_e, err = b.ShareFile("file1", "e")
	if err != nil {
		t.Error("b failed to share file with e")
		return
	}
	err = d.ReceiveFile("file1", "b", magic_string_d)
	if err != nil {
		t.Error("d failed to recieve file1 from b")
		return
	}
	err = e.ReceiveFile("file1", "b", magic_string_e)
	if err != nil {
		t.Error("e failed to recieve file1 from b")
		return
	}
	magic_string_f, err = e.ShareFile("file1", "f")
	if err != nil {
		t.Error("e failed to share file with f")
		return
	}
	err = f.ReceiveFile("file1", "e", magic_string_f)
	if err != nil {
		t.Error("f failed to recieve file1 from e")
		return
	}
	vb, err := b.LoadFile("file1")
	if err != nil {
		t.Error("b failed to load file1")
		return
	}
	vc, err := c.LoadFile("file1")
	if err != nil {
		t.Error("c failed to load file1")
		return
	}
	vd, err := d.LoadFile("file1")
	if err != nil {
		t.Error("d failed to load file1")
		return
	}
	ve, err := e.LoadFile("file1")
	if err != nil {
		t.Error("e failed to load file1")
		return
	}
	vf, err := f.LoadFile("file1")
	if err != nil {
		t.Error("f failed to load file1")
		return
	}
	if !reflect.DeepEqual(va, vb) {
		t.Error("Shared file is not the same (va, vb)", va, vb)
		return
	}
	if !reflect.DeepEqual(va, vc) {
		t.Error("Shared file is not the same (va, vc)", va, vc)
		return
	}
	if !reflect.DeepEqual(va, vd) {
		t.Error("Shared file is not the same (va, vd)", va, vd)
		return
	}
	if !reflect.DeepEqual(va, ve) {
		t.Error("Shared file is not the same (va, ve)", va, ve)
		return
	}
	if !reflect.DeepEqual(va, vf) {
		t.Error("Shared file is not the same (va, vf)", va, vf)
		return
	}
	err = a.RevokeFile("file1", "b")
	if err != nil {
		t.Error("a cannot revoke file")
		return
	}
	_, err = b.LoadFile("file1")
	if err == nil {
		t.Error("b can still load file")
		return
	}
	_, err = d.LoadFile("file1")
	if err == nil {
		t.Error("d can still load file")
		return
	}
	_, err = e.LoadFile("file1")
	if err == nil {
		t.Error("e can still load file")
		return
	}
	_, err = f.LoadFile("file1")
	if err == nil {
		t.Error("f can still load file")
		return
	}
	vc, err = c.LoadFile("file1")
	if err != nil {
		t.Error("c cannot load file")
		return
	}
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a cannot load file")
		return
	}
	err = a.RevokeFile("file1", "c")
	if err != nil {
		t.Error("a failed to revoke c")
		return
	}
	_, err = c.LoadFile("file1")
	if err == nil {
		t.Error("c can still load file")
		return
	}
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a cannot load file")
		return
	}
	v2 := []byte("line2")
	err = f.AppendFile("file1", v2)
	va2, err := a.LoadFile("file1")
	if reflect.DeepEqual(append(va, v2...), va2) {
		t.Error("revoked user can change file")
		return
	}
}

func TestShare11(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	va := []byte("version1")
	a.StoreFile("file1", va)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a's file1 stored incorrectly")
		return
	}
	var magic_string_b string
	magic_string_b, err = a.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share file1 with b")
		return
	}
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b failed to recieve file from a")
		return
	}
	vb, err := b.LoadFile("file1")
	if err != nil {
		t.Error("b's file could not be loaded")
		return
	}
	va2 := []byte("version2")
	a.StoreFile("file1", va2)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a could not load file after restoring")
		return
	}
	if !reflect.DeepEqual(va, va2) {
		t.Error("Restore did not work.")
		return
	}
	vb, err = b.LoadFile("file1")
	if err != nil {
		t.Error("b could not laod file after a restored")
		return
	}
	if !reflect.DeepEqual(va, vb) {
		t.Error("Shared file is not the same (va, vb)", va, vb)
		return
	}
}

func TestShare12(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	_, err = InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	magic_string, err := a.ShareFile("file1", "b")
	if err == nil || magic_string != "" {
		t.Error("Could share file that doesnt exist")
		return
	}
}

func TestShare13(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	filename := randomString(4800)
	v1 := []byte("file1")
	a.StoreFile(filename, v1)
	magic_string, err := a.ShareFile(filename, "b")
	err = b.ReceiveFile(filename, "a", magic_string)
	if err != nil {
		t.Error("b could not recieve file")
		return
	}
}

func TestShare14(t *testing.T) {
	clear()
	a, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	b, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	c, err := InitUser("c", "cc")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	d, err := InitUser("d", "dd")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	e, err := InitUser("e", "ee")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	f, err := InitUser("f", "ff")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	v1 := []byte("file1")
	a.StoreFile("file1", v1)
	magic_string_b, err := a.ShareFile("file1", "b")
	err = b.ReceiveFile("file1", "a", magic_string_b)
	if err != nil {
		t.Error("b could not recieve file")
		return
	}
	magic_string_c, err := a.ShareFile("file1", "c")
	err = c.ReceiveFile("file1", "a", magic_string_c)
	if err != nil {
		t.Error("c could not recieve file")
		return
	}
	magic_string_d, err := b.ShareFile("file1", "d")
	err = d.ReceiveFile("file1", "b", magic_string_d)
	if err != nil {
		t.Error("d could not recieve file")
		return
	}
	magic_string_e, err := b.ShareFile("file1", "e")
	err = e.ReceiveFile("file1", "b", magic_string_e)
	if err != nil {
		t.Error("e could not recieve file")
		return
	}
	magic_string_f, err := c.ShareFile("file1", "f")
	err = f.ReceiveFile("file1", "c", magic_string_f)
	if err != nil {
		t.Error("f could not recieve file")
		return
	}
	ve := []byte("line2e")
	err = e.AppendFile("file1", ve)
	if err != nil {
		t.Error("e could not append")
		return
	}
	va, err := a.LoadFile("file1")
	if err != nil {
		t.Error("a could not load file")
		return
	}
	if !reflect.DeepEqual(va, append(v1, ve...)) {
		t.Error("Appended file is not the same (va, ve)", va, ve)
		return
	}
	vb, err := b.LoadFile("file1")
	if err != nil {
		t.Error("b could not load file")
		return
	}
	if !reflect.DeepEqual(vb, append(v1, ve...)) {
		t.Error("Appended file is not the same (vb, ve)", vb, ve)
		return
	}
	vc, err := c.LoadFile("file1")
	if err != nil {
		t.Error("c could not load file")
		return
	}
	if !reflect.DeepEqual(vc, append(v1, ve...)) {
		t.Error("Appended file is not the same (vc, ve)", vc, ve)
		return
	}
	vd, err := d.LoadFile("file1")
	if err != nil {
		t.Error("d could not load file")
		return
	}
	if !reflect.DeepEqual(vd, append(v1, ve...)) {
		t.Error("Appended file is not the same (vd, ve)", vd, ve)
		return
	}
	vf, err := f.LoadFile("file1")
	if err != nil {
		t.Error("f could not load file")
		return
	}
	if !reflect.DeepEqual(vf, append(v1, ve...)) {
		t.Error("Appended file is not the same (vf, ve)", vf, ve)
		return
	}
	v3 := []byte("Version3")
	f.StoreFile("file1", v3)
	va, err = a.LoadFile("file1")
	if err != nil {
		t.Error("a could not load file")
		return
	}
	if !reflect.DeepEqual(va, v3) {
		t.Error("Appended file is not the same (va, v3)", va, v3)
		return
	}
	vb, err = b.LoadFile("file1")
	if err != nil {
		t.Error("b could not load file")
		return
	}
	if !reflect.DeepEqual(vb, v3) {
		t.Error("Appended file is not the same (vb, v3)", vb, v3)
		return
	}
	vc, err = c.LoadFile("file1")
	if err != nil {
		t.Error("c could not load file")
		return
	}
	if !reflect.DeepEqual(vc, v3) {
		t.Error("Appended file is not the same (vc, v3)", vc, v3)
		return
	}
	vd, err = d.LoadFile("file1")
	if err != nil {
		t.Error("d could not load file")
		return
	}
	if !reflect.DeepEqual(vd, v3) {
		t.Error("Appended file is not the same (vd, ve)", vd, v3)
		return
	}
	ve, err = e.LoadFile("file1")
	if err != nil {
		t.Error("f could not load file")
		return
	}
	if !reflect.DeepEqual(ve, v3) {
		t.Error("Appended file is not the same (ve, v3)", ve, v3)
		return
	}

}

func TestShare15(t *testing.T) {
	clear()
	a1, err := InitUser("a", "aa")
	if err != nil {
		t.Error("Failed to initialize a")
		return
	}
	a2, err := GetUser("a", "aa")
	if err != nil {
		t.Error("could not get second a")
		return
	}
	b1, err := InitUser("b", "bb")
	if err != nil {
		t.Error("Failed to initialize b")
		return
	}
	b2, err := GetUser("b", "bb")
	if err != nil {
		t.Error("Faield to get second b")
		return
	}
	v1 := []byte("file1")
	a1.StoreFile("file1", v1)
	if err != nil {
		t.Error("a failed to store file")
		return
	}
	magic_string, err := a1.ShareFile("file1", "b")
	if err != nil {
		t.Error("a failed to share")
		return
	}
	err = b1.ReceiveFile("file1", "a", magic_string)
	if err != nil {
		t.Error("b could not recieve file")
		return
	}
	vb2, err := b2.LoadFile("file1")
	if err != nil {
		t.Error("b2 could not load file1")
		return
	}
	vb, err := b1.LoadFile("file1")
	if !reflect.DeepEqual(vb, vb2) {
		t.Error("vb and vb2 file1 is not the same")
		return
	}
	va2, err := a2.LoadFile("file1")
	if !reflect.DeepEqual(vb, va2) {
		t.Error("vb and va2 file1 is not the same")
		return
	}
}

// ********************* RANDOM TEST FRAMEWORK ********************* //
// The tests in this section are not designed to fail per se, but are
// instead designed to try and get the system to error in as many
// ways as possible

// struct to hold the state of the test
type testState struct {
	alice1 *User
	alice2 *User
	bob    *User
}

// ************************** HELPERS ****************************** //

// begin each randomized test with this sequence of events
func prologue() *testState {
	clear()
	var state testState

	// create two users and get a second instance of one
	state.alice1, _ = InitUser("alice", "foo")
	state.alice2, _ = GetUser("alice", "foo")
	state.bob, _ = InitUser("bob", "bar")

	// store and append some files
	state.alice1.StoreFile("file1", []byte("This is content 1"))
	state.bob.StoreFile("file2", []byte("This is content 2"))
	state.alice2.AppendFile("file1", []byte("This is content 3"))

	// alice shares her file with bob
	magic, _ := state.alice1.ShareFile("file1", "bob")
	state.bob.ReceiveFile("shared", "alice", magic)

	return &state
}

// get a random integer between 0 and n (not including n), n up to 65535
func getRand(n int) int {
	rand1 := userlib.RandomBytes(1)[0]
	rand2 := userlib.RandomBytes(1)[0]
	return (int(rand1)*256 + int(rand2)) % n
}

// perform random action (fourteen possible random actions)
func randAction(state *testState) (*testState, error) {
	var err error

	// out of 100, so we can adjust probabilities
	n := getRand(100)

	if n < 7 {
		// get another instance of alice
		userlib.DebugMsg("get another alice")
		_, err = GetUser("alice", "foo")
	} else if n < 14 {
		// get another instance of bob
		userlib.DebugMsg("get another bob")
		_, err = GetUser("bob", "bar")
	} else if n < 21 {
		// alice1 loads file1
		userlib.DebugMsg("alice1 loads file1")
		_, err = state.alice1.LoadFile("file1")
	} else if n < 28 {
		// alice2 loads file1
		userlib.DebugMsg("alice2 loads file1")
		_, err = state.alice2.LoadFile("file1")
	} else if n < 35 {
		// bob loads file2
		userlib.DebugMsg("bob loads file2")
		_, err = state.bob.LoadFile("file2")
	} else if n < 42 {
		// bob loads shared
		userlib.DebugMsg("bob loads shared")
		_, err = state.bob.LoadFile("shared")
	} else if n < 50 {
		// alice1 appends to file1
		userlib.DebugMsg("alice1 appends to file1")
		err = state.alice1.AppendFile("file1", []byte("This is some more content"))
	} else if n < 57 {
		// alice2 appends to file1
		userlib.DebugMsg("alice2 appends to file1")
		err = state.alice2.AppendFile("file1", []byte("This is some more content"))
	} else if n < 64 {
		// bob appends to file2
		userlib.DebugMsg("bob appends to file2")
		err = state.bob.AppendFile("file2", []byte("This is some more content"))
	} else if n < 71 {
		// bob appends to shared
		userlib.DebugMsg("bob appends to shared")
		err = state.bob.AppendFile("shared", []byte("This is some more contennt"))
	} else if n < 78 {
		// alice stores another file, which is then immediately received by bob
		userlib.DebugMsg("alice stores another file, which is then immediately received by bob")
		fname := randomString(100)
		state.alice1.StoreFile(fname, []byte("This is a new file"))
		magic, err := state.alice1.ShareFile(fname, "bob")
		if err != nil {
			return state, err
		}
		err = state.bob.ReceiveFile(fname, "alice", magic)
	} else if n < 85 {
		// bob stores another file, which is then immediately received by alice
		userlib.DebugMsg("bob stores another file, which is then immediately received by alice")
		fname := randomString(100)
		state.bob.StoreFile(fname, []byte("This is a new file"))
		magic, err := state.bob.ShareFile(fname, "alice")
		if err != nil {
			return state, err
		}
		err = state.alice2.ReceiveFile(fname, "bob", magic)
	} else if n < 92 {
		// alice1 revokes bob's privileges to her file
		userlib.DebugMsg("alice1 revokes bob's privileges to her file")
		err = state.alice1.RevokeFile("file1", "bob")
		// if nothing errored, reshare file to bob
		if err == nil {
			magic, err := state.alice1.ShareFile("file1", "bob")
			if err != nil {
				return state, err
			}
			err = state.bob.ReceiveFile("shared", "alice", magic)
		}
	} else {
		// alice2 revokes bob's privileges to her file
		userlib.DebugMsg("alice2 revokes bob's privileges to her file")
		err = state.alice2.RevokeFile("file1", "bob")
		// if nothing errored, reshare file to bob
		if err == nil {
			magic, err := state.alice2.ShareFile("file1", "bob")
			if err != nil {
				return state, err
			}
			err = state.bob.ReceiveFile("shared", "alice", magic)
		}
	}

	// return the error if one arose
	if err != nil {
		userlib.DebugMsg("WE TRIPPED AN ERROR!")
		return nil, err
	} else {
		return state, nil
	}
}

// // ************************ RANDOM TESTS *************************** //

func TestChangeSingleBit(t *testing.T) {
	var state *testState
	var datastore map[uuid.UUID][]byte
	var k uuid.UUID
	var v []byte
	var n, rand int
	var err error

	// // repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()
		datastore = userlib.DatastoreGetMap()

		// pick a random entry in the datastore
		n = len(datastore)
		rand = getRand(n)
		for k, v = range datastore {
			if rand != 0 {
				rand--
			} else {
				break
			}
		}
		n = len(v) // pick a byte in the value at that location
		rand = getRand(n)
		v[rand] = v[rand] ^ (1 << uint8(getRand(8))) // flip the bit
		userlib.DatastoreSet(k, v)

		// try 25 times to get it to error after flipping the bit
		for j := 0; j < 25; j++ {
			_, err = randAction(state)
			if err != nil {
				break
			}
		}
	}
}

func TestChangeManyBits(t *testing.T) {
	var state *testState
	var datastore map[uuid.UUID][]byte
	var k uuid.UUID
	var v []byte
	var n, rand int
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()
		datastore = userlib.DatastoreGetMap()

		// pick a random entry in the datastore
		n = len(datastore)
		rand = getRand(n)
		for k, v = range datastore {
			if rand != 0 {
				rand--
			} else {
				break
			}
		}
		// flip 256 bits in a particular value in the datstore
		for k := 0; k < 256; k++ {
			// pick a random byte in the value
			n = len(v)
			rand = getRand(n)
			v[rand] = v[rand] ^ (1 << uint8(getRand(8))) // flip this bit
		}
		userlib.DatastoreSet(k, v)

		// try 25 times to get it to error after flipping 256 bits
		for j := 0; j < 25; j++ {
			state, err = randAction(state)
			if err != nil {
				break
			}
		}
	}
}

func TestDeleteEntries(t *testing.T) {
	var state *testState
	var datastore map[uuid.UUID][]byte
	var k uuid.UUID
	var n, rand int
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()
		datastore = userlib.DatastoreGetMap()

		// pick a random entry in the datastore to delete
		n = len(datastore)
		rand = getRand(n)
		for k, _ = range datastore {
			if rand != 0 {
				rand--
			} else {
				break
			}
		}
		userlib.DatastoreDelete(k)

		// try 25 times to get it to error after deleting entry
		for j := 0; j < 25; j++ {
			state, err = randAction(state)
			if err != nil {
				break
			}
		}
	}
}

func TestOverwriteEntries(t *testing.T) {
	var state *testState
	var datastore map[uuid.UUID][]byte
	var k uuid.UUID
	var n, rand int
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()
		datastore = userlib.DatastoreGetMap()

		// pick a random entry in the datastore to overwrite with garbage
		n = len(datastore)
		rand = getRand(n)
		for k, _ = range datastore {
			if rand != 0 {
				rand--
			} else {
				break
			}
		}
		userlib.DatastoreSet(k, userlib.RandomBytes(getRand(793)))

		// try 25 times to get it to error after overwriting entry
		for j := 0; j < 25; j++ {
			state, err = randAction(state)
			if err != nil {
				break
			}
		}
	}
}

func TestSwapEntries(t *testing.T) {
	var state *testState
	var datastore map[uuid.UUID][]byte
	var k1, k2 uuid.UUID
	var v1, v2 []byte
	var n, rand int
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()
		datastore = userlib.DatastoreGetMap()

		// pick two random entries in the datastore
		n = len(datastore)
		rand = getRand(n)
		for k1, v1 = range datastore {
			if rand != 0 {
				rand--
			} else {
				break
			}
		}
		rand = getRand(n)
		for k2, v2 = range datastore {
			if rand != 0 {
				rand--
			} else {
				break
			}
		}
		// swap the values associated with the two keys
		userlib.DatastoreSet(k1, v2)
		userlib.DatastoreSet(k2, v1)

		// try 25 times to get it to error after swapping entries
		for j := 0; j < 25; j++ {
			state, err = randAction(state)
			if err != nil {
				break
			}
		}
	}
}

func TestMagicStringFlipSingleBit(t *testing.T) {
	var state *testState
	var magic string
	var magicAsByte []byte
	var n, rand int
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()

		// make a new file and obtain magic string when sharing with bob
		state.alice1.StoreFile("testfile", []byte("This is another test file"))
		magic, _ = state.alice1.ShareFile("testfile", "bob")
		magicAsByte = []byte(magic)

		// pick a random byte in magic string
		n = len(magicAsByte)
		rand = getRand(n)
		magicAsByte[rand] = magicAsByte[rand] ^ (1 << uint8(getRand(8))) // flip this bit

		// Bob tries to receive with altered magic string
		err = state.bob.ReceiveFile("another_shared", "alice", string(magicAsByte))
		if err == nil {
			t.Error("bob was able to receive file with a flipped bit in magic string")
		}
	}
}

func TestMagicStringFlipManyBits(t *testing.T) {
	var state *testState
	var magic string
	var magicAsByte []byte
	var n, rand int
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()

		// make a new file and obtain magic string when sharing with bob
		state.alice1.StoreFile("testfile", []byte("This is another test file"))
		magic, _ = state.alice1.ShareFile("testfile", "bob")
		magicAsByte = []byte(magic)

		// flip 256 random bits in the magic string
		for i := 0; i < 256; i++ {
			// pick a random byte in magic string
			n = len(magicAsByte)
			rand = getRand(n)
			magicAsByte[rand] = magicAsByte[rand] ^ (1 << uint8(getRand(8))) // flip this bit
		}

		// Bob tries to receive with altered magic string
		err = state.bob.ReceiveFile("another_shared", "alice", string(magicAsByte))
		if err == nil {
			t.Error("bob was able to receive file with many flipped bits in magic string")
		}
	}
}

func TestMagicStringLargeGarbage(t *testing.T) {
	var state *testState
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()

		// Bob tries to receive with garbage magic string
		err = state.bob.ReceiveFile("another_shared", "alice", string(userlib.RandomBytes(getRand(793))))
		if err == nil {
			t.Error("bob was able to receive file with garbage magic string")
		}
	}
}
func TestMagicStringSmallGarbage(t *testing.T) {
	var state *testState
	var err error

	// repeat this tests this number of times
	for i := 0; i < 200; i++ {
		state = prologue()

		// Bob tries to receive with garbage magic string
		err = state.bob.ReceiveFile("another_shared", "alice", string(userlib.RandomBytes(getRand(101))))
		if err == nil {
			t.Error("bob was able to receive file with garbage magic string")
		}
	}
}
