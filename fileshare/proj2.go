package proj2

// CS 161 Project 2 Fall 2020
// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder. We will be very upset.

import (
	// You neet to add with
	// go get github.com/cs161-staff/userlib

	"github.com/cs161-staff/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging, etc...
	"encoding/hex"

	// UUIDs are generated right based on the cryptographic PRNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys.
	"strings"

	// Want to import errors.
	"errors"

	// Optional. You can remove the "_" there, but please do not touch
	// anything else within the import bracket.
	_ "strconv"
	// if you are looking for fmt, we don't give you fmt, but you can use userlib.DebugMsg.
	// see someUsefulThings() below:
)

// ******************************************* STRUCTS *********************************** //

// The structure definition for a user record
type User struct {
	Username string
	password string
	PKEDec   userlib.PKEDecKey
	DSSign   userlib.DSSignKey
	UUIDMap  map[string]uuid.UUID
	OwnerMap map[string]string
}

// The structure definition for a File Sentinel record
type Sentinel struct {
	MetadataKeyMap map[string][]byte
	Lock           []byte
	MetadataUUID   uuid.UUID
}

// The structure definition for a node in the AccessTree
type AccessNode struct {
	Username string
	ShareTo  []*AccessNode
}

// The structure definition for a File Metadata record
type Metadata struct {
	MetadataUUID uuid.UUID
	AccessMap    map[string][]string
	LastModified string
	Owner        string
	TextEncKey   []byte
	TextMACKey   []byte
	TextList     []uuid.UUID
}

// The structure definition for a File Text Block record
type Text struct {
	TextUUID uuid.UUID
	Data     []byte
}

// The structure definition for a magic string
type magicStringStruct struct {
	SentinelUUID uuid.UUID
	Owner        string
}

// **************************************** DEBUG HELPERS ******************************* //

// Pretty-prints JSON in the terminal so it's easier to read
func prettyPrint(obj interface{}) {
	prettyJSON, _ := json.MarshalIndent(obj, "", "\t")
	userlib.DebugMsg(string(prettyJSON))
}

// Prints out bytes
func printBytes(bytes []byte) {
	userlib.DebugMsg(hex.EncodeToString([]byte(bytes)))
}

// ************************************ LOW-LEVEL UTILITIES ****************************** //

// Computes the bitwise NOT of all bytes in the input slice
func negateBytes(data []byte) []byte {
	var ret = make([]byte, len(data))
	for i := range data {
		ret[i] = ^data[i]
	}
	return ret
}

// Pads data to a multiple of the block cipher blocklen (16 bytes)
func pad(data []byte) []byte {
	var bytesToPad = userlib.AESBlockSize - (len(data) % userlib.AESBlockSize) // AESBlockSize = 16
	var ret = make([]byte, len(data))
	copy(ret, data)

	// don't pad zero bytes (can't tell between good data and padding)
	if bytesToPad == 0 {
		bytesToPad = userlib.AESBlockSize
	}
	for i := 0; i < bytesToPad; i++ {
		ret = append(ret, byte(bytesToPad))
	}
	return ret
}

// Unpads data from block cipher blocklen (16 bytes) according to padding scheme
func unpad(data []byte) []byte {
	var bytesToRemove = int(data[len(data)-1])
	return data[:len(data)-bytesToRemove]
}

// Slow hash function that does SHA512 10,000 times
func slowHash(data []byte) (ret [64]byte) {
	const count = 10000 // number of times to do the hash

	// compute hash count number of times
	ret = userlib.Hash(data)
	for i := 0; i < count; i++ {
		ret = userlib.Hash(ret[:])
	}
	return ret
}

// Computes UUID of User struct associated with given username and password
func getUserUUID(username string, password string) (ret uuid.UUID) {
	var uuidHash = slowHash([]byte(username + "-" + password))
	var uuidSlice = uuidHash[:]
	for i := range ret {
		ret[i] = uuidSlice[i]
	}
	return ret
}

// Computes symmetric encryption and authentication keys for User struct
// with given username, password, and salt
func getUserKeys(username string, password string, salt []byte) (symEnc []byte, symMAC []byte) {
	symEnc = userlib.Argon2Key([]byte(password), salt, uint32(userlib.AESBlockSize))
	symMAC = userlib.Argon2Key([]byte(password), negateBytes(salt), uint32(userlib.AESBlockSize))
	return symEnc, symMAC
}

// Computes the hash of the given filename for a username as a string
func getFilenameHash(filename string, username string) string {
	hash := userlib.Hash([]byte(filename + "-" + username))
	return hex.EncodeToString(hash[:])
}

// Computes the value under the sentinel lock: Hash(file_metadata_UUID || file_metadata_key)
func getSentinelLockVal(metadataUUID uuid.UUID, metadataKey []byte) (val []byte) {
	var uuidSlice = []byte(metadataUUID.String())
	var asArray = userlib.Hash(append(uuidSlice, metadataKey...))
	return asArray[:]
}

// Return true if accessMap has key username
func accessContains(accessMap map[string][]string, username string) (ok bool) {
	if _, ok := accessMap[username]; ok {
		return true
	}
	return false
}

// Return true if a shared with b
func shared(a string, b string, accessMap map[string][]string) (ok bool) {
	if accessContains(accessMap, a) && accessContains(accessMap, b) {
		for _, val := range accessMap[a] {
			if val == b {
				return true
			}
		}
	}
	return false
}

// Remove b from a's accessArray
func removeAccess(accessMap map[string][]string, a string, b string) (newAccessMap map[string][]string, err error) {
	var list []string
	for _, n := range accessMap[a] {
		if n != b {
			list = append(list, n)
		}
	}
	if len(list) == len(accessMap[a]) {
		return accessMap, errors.New(strings.ToTitle("Target not found."))
	}
	for _, n := range accessMap[b] {
		accessMap, err = removeAccess(accessMap, b, n)
		if err != nil {
			return accessMap, err
		}
	}
	delete(accessMap, b)
	accessMap[a] = list
	newAccessMap = accessMap
	return newAccessMap, nil

}

// ************************************** LARGE SUBROUTINES ****************************** //

// Posts a general struct to datastore
// Takes care of marshalling, padding, (salting), encrypting, authenticating
// Authenticates with symMAC if sign = false, otherwise authenticates with DSSign
func postStruct(obj interface{}, newUUID uuid.UUID, symEnc []byte, symMAC []byte, salt []byte, usr User, sign bool) (err error) {
	// variable declarations
	var plaintext, ciphertext, auth, val []byte

	// get plaintext and pad
	plaintext, err = json.Marshal(obj)
	if err != nil {
		userlib.DebugMsg("error when marshalling object")
		return err
	}
	plaintext = pad(plaintext)

	// encrypt, authenticate, and store data into datastore
	ciphertext = userlib.SymEnc(symEnc, userlib.RandomBytes(userlib.AESBlockSize), plaintext)
	if !sign && symMAC != nil {
		auth, err = userlib.HMACEval(symMAC, ciphertext)
		if err != nil {
			userlib.DebugMsg("error computing HMAC of object ciphertext")
			return err
		}
	} else {
		auth, err = userlib.DSSign(usr.DSSign, ciphertext)
		if err != nil {
			userlib.DebugMsg("error signing object ciphertext")
			return err
		}
	}
	if salt != nil {
		val = append(auth, salt...)
		val = append(val, ciphertext...)
	} else {
		val = append(auth, ciphertext...)
	}
	userlib.DatastoreSet(newUUID, val)

	return nil
}

// Posts the User struct with username and password to datastore
func postUser(usr User, username string, password string) (err error) {
	// variable declarations
	var usrUUID uuid.UUID
	var salt, symEnc, symMAC []byte

	// generate and compute necessary values
	salt = userlib.RandomBytes(40)
	usrUUID = getUserUUID(username, password)
	symEnc, symMAC = getUserKeys(username, password, salt)

	// post to datastore
	err = postStruct(usr, usrUUID, symEnc, symMAC, salt, usr, false)
	if err != nil {
		userlib.DebugMsg("error when posting user struct of " + usr.Username)
		return err
	}
	return nil
}

// Post the given Metadata struct to datastore
func postMetadata(usr User, metadata Metadata, metadataKey []byte) (err error) {
	err = postStruct(metadata, metadata.MetadataUUID, metadataKey, nil, nil, usr, true)
	if err != nil {
		userlib.DebugMsg("error when posting metadata struct for file")
		return err
	}
	return nil
}

// Post the given Text struct to datastore
func postText(usr User, text Text, metadata Metadata) (err error) {
	err = postStruct(text, text.TextUUID, metadata.TextEncKey, metadata.TextMACKey, nil, usr, false)
	if err != nil {
		userlib.DebugMsg("error when posting text struct for file")
		return err
	}
	return nil
}

// Verifies that the given User has access to the file with the given filename
func verifyFileAccessHelper(usr User, filenameHash string) (metadata Metadata, metadataKey []byte, err error) {
	// variable declarations
	var ok bool
	var sentinel Sentinel
	var ownerDSPub, lastModDSPub userlib.DSVerifyKey
	var sentinelUUID, metadataUUID uuid.UUID
	var plaintext, ciphertext, sig, val []byte

	// make sure that filenameHash is in usr.UUIDMap
	sentinelUUID, ok = usr.UUIDMap[filenameHash]
	if !ok {
		userlib.DebugMsg("cannot find UUID of file sentinel!")
		return metadata, nil, errors.New(strings.ToTitle("user cannot find UUID of file sentinel!"))
	}

	// make sure owner has a posted DS public key
	ownerDSPub, ok = userlib.KeystoreGet(usr.OwnerMap[filenameHash] + "-DSVerifyKey")
	if !ok {
		userlib.DebugMsg(usr.OwnerMap[filenameHash] + " does not have a posted DS Verify Key!")
		return metadata, nil, errors.New(strings.ToTitle("cannot find file owner's DS Verify key!"))
	}

	// retrieve the file sentinel
	plaintext, ok = userlib.DatastoreGet(sentinelUUID)
	if !ok {
		userlib.DebugMsg("file sentinel is not at recorded UUID")
		return metadata, nil, errors.New(strings.ToTitle("file sentinel is not at recorded UUID"))
	}
	err = json.Unmarshal(plaintext, &sentinel)
	if err != nil {
		userlib.DebugMsg("error unmarshalling file sentinel")
		return metadata, nil, err
	}

	// get and decrypt the file metadata key
	ciphertext, ok = sentinel.MetadataKeyMap[usr.Username]
	if !ok {
		userlib.DebugMsg("no encrypted key entry in file sentinel for this user")
		return metadata, nil, errors.New(strings.ToTitle("no ecnrypted key entry in file sentinel for this user"))
	}
	metadataKey, err = userlib.PKEDec(usr.PKEDec, ciphertext)
	if err != nil {
		userlib.DebugMsg("error decrypting key entry in file sentinel for this user")
		return metadata, nil, err
	}

	// verify the file sentinel lock
	metadataUUID = sentinel.MetadataUUID
	val = getSentinelLockVal(metadataUUID, metadataKey)
	if err != nil {
		userlib.DebugMsg("error computing sentinel lock value")
		return metadata, nil, err
	}
	err = userlib.DSVerify(ownerDSPub, val, sentinel.Lock)
	if err != nil {
		userlib.DebugMsg("could not verify file sentinel lock; data is corrupted!")
		return metadata, nil, err
	}

	// get and split file metadata into components
	val, ok = userlib.DatastoreGet(metadataUUID)
	if !ok {
		userlib.DebugMsg("file metadata is not at recorded UUID")
		return metadata, nil, errors.New(strings.ToTitle("file metadata is not at recorded UUID"))
	}
	if len(val) < userlib.RSAKeySize/8 {
		userlib.DebugMsg("issue with retrieving file metadata, incorrect array len")
		return metadata, nil, errors.New(strings.ToTitle("file metadata val is not correct length"))
	}
	sig = val[:userlib.RSAKeySize/8]
	ciphertext = val[userlib.RSAKeySize/8:]

	// decrypt then verify (this is in the wrong order but we don't think it's an issue because of other security layers above)
	if len(ciphertext)%userlib.AESBlockSize != 0 {
		userlib.DebugMsg("ciphertext is not a multiple of the block size!")
		return metadata, nil, errors.New(strings.ToTitle("ciphertext is not a multiple of the block size!"))
	}
	plaintext = userlib.SymDec(metadataKey, ciphertext)
	plaintext = unpad(plaintext)
	err = json.Unmarshal(plaintext, &metadata)
	if err != nil {
		userlib.DebugMsg("error unmarshalling file metadata")
		return metadata, nil, err
	}
	lastModDSPub, ok = userlib.KeystoreGet(metadata.LastModified + "-DSVerifyKey")
	if !ok {
		userlib.DebugMsg(metadata.LastModified + " does not have a posted DS Verify Key!")
		return metadata, nil, errors.New(strings.ToTitle("cannot find last modified's DS Verify key!"))
	}
	err = userlib.DSVerify(lastModDSPub, ciphertext, sig)
	if err != nil {
		userlib.DebugMsg("signature on file metadata did not match; data has been corrupted")
		return metadata, nil, err
	}
	if metadata.MetadataUUID != metadataUUID {
		userlib.DebugMsg("metadata UUID's do not match!")
		return metadata, nil, errors.New(strings.ToTitle("metadata UUID's do not match!"))
	}

	return metadata, metadataKey, nil
}

// Calls the helper to run through checks; on failure, deletes entries in UUIDMap and OwnerMap
// On success, returns metadata and metadata key
func verifyFileAccess(usr User, filenameHash string) (metadata Metadata, metadataKey []byte, err error) {
	metadata, metadataKey, err = verifyFileAccessHelper(usr, filenameHash)
	if err != nil {
		delete(usr.UUIDMap, filenameHash)
		delete(usr.OwnerMap, filenameHash)
		userlib.DebugMsg("could not verify that " + usr.Username + " has access to file")
		return metadata, nil, errors.New(strings.ToTitle("could not verify that user has access to file"))
	}
	return metadata, metadataKey, nil
}

// **************************************** API FUNCTIONS ******************************** //

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the password has strong entropy, EXCEPT
// the attackers may possess a precomputed tables containing
// hashes of common passwords downloaded from the internet.
func InitUser(username string, password string) (userdataptr *User, err error) {
	// variable declarations
	var usr User
	var ok bool
	var DSPriv userlib.DSSignKey
	var DSPub userlib.DSVerifyKey
	var PKEPriv userlib.PKEDecKey
	var PKEPub userlib.PKEEncKey

	// check if username has been taken in keystore
	_, ok = userlib.KeystoreGet(username + "-DSVerifyKey")
	if ok {
		userlib.DebugMsg("user " + username + " already exists!")
		return nil, errors.New(strings.ToTitle("user already exists!"))
	}
	_, ok = userlib.KeystoreGet(username + "-PKEEncKey")
	if ok {
		userlib.DebugMsg("user " + username + " already exists!")
		return nil, errors.New(strings.ToTitle("user already exists!"))
	}

	// generate DS key pair and store in usr struct, keystore
	DSPriv, DSPub, err = userlib.DSKeyGen()
	if err != nil {
		userlib.DebugMsg("error when generating DS key for " + username)
		return nil, err
	}
	err = userlib.KeystoreSet(username+"-DSVerifyKey", DSPub)
	if err != nil {
		userlib.DebugMsg("error when setting DS Pub key of " + username + " to Keystore")
		return nil, err
	}
	usr.DSSign = DSPriv

	// generate PKE key pair and store in usr struct, keystore
	PKEPub, PKEPriv, err = userlib.PKEKeyGen()
	if err != nil {
		userlib.DebugMsg("error when generating PKE key for " + username)
		return nil, err
	}
	err = userlib.KeystoreSet(username+"-PKEEncKey", PKEPub)
	if err != nil {
		userlib.DebugMsg("error when setting PKE Enc key of " + username + " to Keystore")
		return nil, err
	}
	usr.PKEDec = PKEPriv

	// fill out remaining fields in usr
	usr.UUIDMap = make(map[string]uuid.UUID)
	usr.OwnerMap = make(map[string]string)
	usr.Username = username
	usr.password = password

	// post usr to datastore
	err = postUser(usr, username, password)
	if err != nil {
		userlib.DebugMsg("error when submitting user struct of " + username + " to Datastore")
		return nil, err
	}

	return &usr, nil
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	// variable declarations
	var usr User
	var ok bool
	var usrUUID uuid.UUID
	var salt, symEnc, symMAC, plaintext, ciphertext, authStore, authCompute, val []byte

	// check if this user has keys in keystore
	_, ok = userlib.KeystoreGet(username + "-DSVerifyKey")
	if !ok {
		userlib.DebugMsg(username + " does not exist! Could not find associated DS key")
		return nil, errors.New(strings.ToTitle("user does not exist!"))
	}
	_, ok = userlib.KeystoreGet(username + "-PKEEncKey")
	if !ok {
		userlib.DebugMsg(username + "does not exist! Could not find associated PKE Enc key")
		return nil, errors.New(strings.ToTitle("user does not exist!"))
	}

	// check that UUID (as a function of username and password) exists in datastore
	usrUUID = getUserUUID(username, password)
	val, ok = userlib.DatastoreGet(usrUUID)
	if !ok {
		userlib.DebugMsg("Incorrect password " + password + " or data corrupted")
		return nil, errors.New(strings.ToTitle("incorrect password or data corrupted"))
	}

	// split val into its component parts
	if len(val) < userlib.HashSize+40 {
		userlib.DebugMsg("value for user struct was corrupted in DataStore (not enough info)")
		return nil, errors.New(strings.ToTitle("value for user struct was corrupted in DataStore (not enough info)"))
	}
	authStore = val[:userlib.HashSize]
	salt = val[userlib.HashSize : userlib.HashSize+40]
	ciphertext = val[userlib.HashSize+40:]

	// compute symmetric keys based on username, salt for encryption, authentication
	symEnc, symMAC = getUserKeys(username, password, salt)

	// verify, decrypt, and unmarshal val
	authCompute, err = userlib.HMACEval(symMAC, ciphertext)
	if err != nil {
		userlib.DebugMsg("error computing HMAC of ciphertext for " + username)
		return nil, err
	}
	if !userlib.HMACEqual(authStore, authCompute) {
		userlib.DebugMsg("cannot verify User struct for " + username)
		return nil, errors.New(strings.ToTitle("cannot verify User struct"))
	}
	if len(ciphertext)%userlib.AESBlockSize != 0 {
		userlib.DebugMsg("ciphertext is not a multiple of the block size!")
		return nil, errors.New(strings.ToTitle("ciphertext is not a multiple of the block size!"))
	}
	plaintext = userlib.SymDec(symEnc, ciphertext)
	plaintext = unpad(plaintext)
	err = json.Unmarshal(plaintext, &usr)
	if err != nil {
		userlib.DebugMsg("error unmarshalling user struct for " + username)
		return nil, errors.New(strings.ToTitle("cannot unmarshal User struct"))
	}

	// check that username matches username provided
	if usr.Username != username {
		userlib.DebugMsg("username mismatch, something fishy is going on...")
		return nil, errors.New(strings.ToTitle("username mismatch"))
	}

	// fill in private field
	usr.password = password

	return &usr, nil
}

// This stores a file in the datastore.
//
// The plaintext of the filename + the plaintext and length of the filename
// should NOT be revealed to the datastore!
func (usr *User) StoreFile(filename string, data []byte) {
	// variable declarations
	var sentinel Sentinel
	var metadata Metadata
	var text Text
	var ownerNode AccessNode
	var PKEPub userlib.PKEEncKey
	var err error
	var filenameHash string
	var ok, isNewFile bool
	var sentinelUUID, metadataUUID, textUUID uuid.UUID
	var metadataKey, symEnc, symMAC, plaintext, val []byte
	var shared []string

	// first, determine if we are storing a completely new file or updating an existing one
	usr, err = GetUser(usr.Username, usr.password)
	if err != nil {
		userlib.DebugMsg("Could not update User struct!")
		return
	}
	filenameHash = getFilenameHash(filename, usr.Username)
	sentinelUUID, ok = usr.UUIDMap[filenameHash]
	if !ok {
		isNewFile = true
	} else {
		isNewFile = false
	}

	// generate, compute, and retrieve several values
	textUUID = uuid.New()
	symEnc = userlib.RandomBytes(userlib.AESBlockSize)
	symMAC = userlib.RandomBytes(userlib.AESBlockSize)
	PKEPub, ok = userlib.KeystoreGet(usr.Username + "-PKEEncKey")
	if !ok {
		userlib.DebugMsg(usr.Username + " does not have a PKE Public Key in Keystore!")
		return
	}

	// if new file, create new file sentinel and file metadata
	// if not new file, retrieve file metadata
	if isNewFile {
		// generate more values for new files
		sentinelUUID = uuid.New()
		metadataUUID = uuid.New()
		metadataKey = userlib.RandomBytes(userlib.AESBlockSize)

		// update maps in user struct
		usr.UUIDMap[filenameHash] = sentinelUUID
		usr.OwnerMap[filenameHash] = usr.Username

		// construct file sentinel
		sentinel.MetadataKeyMap = make(map[string][]byte)
		sentinel.MetadataKeyMap[usr.Username], err = userlib.PKEEnc(PKEPub, metadataKey)
		if err != nil {
			userlib.DebugMsg("error RSA-encrypting file metadata key")
			return
		}
		val = getSentinelLockVal(metadataUUID, metadataKey)
		sentinel.Lock, err = userlib.DSSign(usr.DSSign, val)
		if err != nil {
			userlib.DebugMsg("error computing sentinel lock value")
			return
		}
		sentinel.MetadataUUID = metadataUUID

		// post file sentinel to userdata
		plaintext, err = json.Marshal(sentinel)
		if err != nil {
			userlib.DebugMsg("error marshalling file sentinel")
			return
		}
		userlib.DatastoreSet(sentinelUUID, plaintext)

		// construct file metadata
		metadata.MetadataUUID = metadataUUID
		metadata.Owner = usr.Username
		metadata.TextEncKey = symEnc
		metadata.TextMACKey = symMAC
		ownerNode.Username = usr.Username
		metadata.AccessMap = make(map[string][]string)
		metadata.AccessMap[usr.Username] = shared
	} else {
		metadata, metadataKey, err = verifyFileAccess(*usr, filenameHash)
		if err != nil {
			userlib.DebugMsg("could not verify that " + usr.Username + " has access to " + filename)
			return
		}
	}

	// update file metadata
	metadata.LastModified = usr.Username
	metadata.TextList = nil
	metadata.TextList = append(metadata.TextList, textUUID)

	// construct text block
	text.TextUUID = textUUID
	text.Data = data

	// post new/updated user struct, file metadata, and text block
	err = postUser(*usr, usr.Username, usr.password)
	if err != nil {
		userlib.DebugMsg("error posting user struct")
	}
	err = postMetadata(*usr, metadata, metadataKey)
	if err != nil {
		userlib.DebugMsg("error posting file metadata")
		return
	}
	err = postText(*usr, text, metadata)
	if err != nil {
		userlib.DebugMsg("error posting text block")
		return
	}
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (usr *User) AppendFile(filename string, data []byte) (err error) {
	// variable declarations
	var metadata Metadata
	var text Text
	var filenameHash string
	var metadataKey []byte

	// verify that this user has access to the given file
	usr, err = GetUser(usr.Username, usr.password)
	if err != nil {
		userlib.DebugMsg("Could not update User struct!")
		return err
	}
	filenameHash = getFilenameHash(filename, usr.Username)
	metadata, metadataKey, err = verifyFileAccess(*usr, filenameHash)
	if err != nil {
		userlib.DebugMsg("could not verify that " + usr.Username + " has access to " + filename)
		return errors.New(strings.ToTitle("could not verify that user has access to file"))
	}

	// construct new text block
	text.TextUUID = uuid.New()
	text.Data = data

	// update metadata
	metadata.LastModified = usr.Username
	metadata.TextList = append(metadata.TextList, text.TextUUID)

	// post updated file metadata and new text block
	err = postMetadata(*usr, metadata, metadataKey)
	if err != nil {
		userlib.DebugMsg("error posting file metadata")
		return err
	}
	err = postText(*usr, text, metadata)
	if err != nil {
		userlib.DebugMsg("error posting text block")
		return err
	}

	return nil
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (usr *User) LoadFile(filename string) (data []byte, err error) {
	// variable declarations
	var metadata Metadata
	var text Text
	var ok bool
	var filenameHash string
	var plaintext, ciphertext, authStore, authCompute, val []byte

	// verify that this user has access to the given file
	usr, err = GetUser(usr.Username, usr.password)
	if err != nil {
		userlib.DebugMsg("Could not update User struct!")
		return nil, err
	}
	filenameHash = getFilenameHash(filename, usr.Username)
	metadata, _, err = verifyFileAccess(*usr, filenameHash)
	if err != nil {
		userlib.DebugMsg("could not verify that " + usr.Username + " has access to " + filename)
		return nil, errors.New(strings.ToTitle("could not verify that user has access to filename"))
	}

	// for each text block in metadata.TextList
	for i := range metadata.TextList {
		// retrieve from datastore and split val into components
		val, ok = userlib.DatastoreGet(metadata.TextList[i])
		if !ok {
			userlib.DebugMsg("text block is not at recorded UUID")
			return nil, errors.New(strings.ToTitle("text block is not at recorded UUID"))
		}
		if len(val) < userlib.HashSize {
			userlib.DebugMsg("text block was corrupted in datastore (not enough info)")
			return nil, errors.New(strings.ToTitle("text block was corrupted in datastore (not enough info)"))
		}
		authStore = val[:userlib.HashSize]
		ciphertext = val[userlib.HashSize:]

		// verify, decrypt, and unmarshal text block
		authCompute, err = userlib.HMACEval(metadata.TextMACKey, ciphertext)
		if err != nil {
			userlib.DebugMsg("error computing HMAC of ciphertext for text block")
			return nil, err
		}
		if !userlib.HMACEqual(authStore, authCompute) {
			userlib.DebugMsg("cannot verify text block for " + filename)
			return nil, errors.New(strings.ToTitle("cannot verify text block"))
		}
		if len(ciphertext)%userlib.AESBlockSize != 0 {
			userlib.DebugMsg("ciphertext is not a multiple of the block size!")
			return nil, errors.New(strings.ToTitle("ciphertext is not a multiple of the block size!"))
		}
		plaintext = userlib.SymDec(metadata.TextEncKey, ciphertext)
		plaintext = unpad(plaintext)
		err = json.Unmarshal(plaintext, &text)
		if err != nil {
			userlib.DebugMsg("error unmarshalling text block")
			return nil, err
		}
		if metadata.TextList[i] != text.TextUUID {
			userlib.DebugMsg("malicious user swapped text blocks!")
			return nil, errors.New(strings.ToTitle("malicious user swapped text blocks!"))
		}

		// concatenate data in this text block to data
		data = append(data, text.Data...)
	}

	return data, nil
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.
func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {

	// variable declarations
	var metadata Metadata
	var ok bool
	var filenameHash, signatureString string
	var recipientPKEEnc userlib.PKEEncKey
	var sentinel Sentinel
	var sentinelUUID uuid.UUID
	var metadataKey, symEnc, signature, plaintext, magicStringPlaintext, magicStringCombined, ciphertext, key []byte
	var magicStringStruc magicStringStruct
	var share []string

	userdata, err = GetUser(userdata.Username, userdata.password)
	if err != nil {
		userlib.DebugMsg("Could not update User struct!")
		return "", err
	}
	// verify user has access to given file
	filenameHash = getFilenameHash(filename, userdata.Username)
	metadata, metadataKey, err = verifyFileAccess(*userdata, filenameHash)
	if err != nil {
		userlib.DebugMsg("could not verify that " + userdata.Username + " has access to " + filename)
		return "", errors.New(strings.ToTitle("could not verify that user has access to file"))
	}

	// check that recipient is valid
	recipientPKEEnc, ok = userlib.KeystoreGet(recipient + "-PKEEncKey")
	if !ok {
		userlib.DebugMsg(recipient + " does not exist! Could not find associated PKE Encryption Key")
		return "", errors.New(strings.ToTitle("recipient does not exist!"))
	}

	// update AccessTree, AccessList, and last modified
	metadata.AccessMap[userdata.Username] = append(metadata.AccessMap[userdata.Username], recipient)
	metadata.AccessMap[recipient] = share
	metadata.LastModified = userdata.Username

	// post metadata
	err = postMetadata(*userdata, metadata, metadataKey)
	if err != nil {
		userlib.DebugMsg("error posting file metadata")
		return "", err
	}

	// update file sentinel
	filenameHash = getFilenameHash(filename, userdata.Username)
	sentinelUUID, ok = userdata.UUIDMap[filenameHash]
	if !ok {
		userlib.DebugMsg("cannot find UUID of file sentinel!")
		return "", errors.New(strings.ToTitle("user cannot find UUID of file sentinel!"))
	}
	plaintext, ok = userlib.DatastoreGet(sentinelUUID)
	if !ok {
		userlib.DebugMsg("file sentinel is not at recorded UUID")
		return "", errors.New(strings.ToTitle("file sentinel is not at recorded UUID"))
	}
	err = json.Unmarshal(plaintext, &sentinel)
	if err != nil {
		userlib.DebugMsg("error unmarshalling file sentinel")
		return "", err
	}
	sentinel.MetadataKeyMap[recipient], err = userlib.PKEEnc(recipientPKEEnc, metadataKey)
	if err != nil {
		userlib.DebugMsg("error RSA-encrypting file metadata key")
		return "", err
	}

	// post file sentinel to userdata
	plaintext, err = json.Marshal(sentinel)
	if err != nil {
		userlib.DebugMsg("error marshalling file sentinel")
		return "", err
	}
	userlib.DatastoreSet(sentinelUUID, plaintext)

	// generate magic string
	magicStringStruc.Owner = metadata.Owner
	magicStringStruc.SentinelUUID = sentinelUUID
	symEnc = userlib.RandomBytes(userlib.AESBlockSize)
	magicStringPlaintext, err = json.Marshal(magicStringStruc)
	if err != nil {
		userlib.DebugMsg("error marshalling magicstring structure")
		return "", err
	}
	magicStringPlaintext = pad(magicStringPlaintext)
	ciphertext = userlib.SymEnc(symEnc, userlib.RandomBytes(userlib.AESBlockSize), magicStringPlaintext)
	key, err = userlib.PKEEnc(recipientPKEEnc, symEnc)
	if err != nil {
		userlib.DebugMsg("Error encrypting symenc key")
		return "", err
	}
	magicStringCombined = append(key, ciphertext...)
	signature, err = userlib.DSSign(userdata.DSSign, magicStringCombined)
	if err != nil {
		userlib.DebugMsg("Error signing")
		return "", err
	}
	signatureString = string(signature)
	magic_string = signatureString + string(magicStringCombined)

	return magic_string, nil
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	// variable declarations
	var senderDSPub userlib.DSVerifyKey
	var ok bool
	var signatureByte, magicStringByte, magic_stringByte, key, plaintext, keyEnc []byte
	var err error
	var magicStringStruc magicStringStruct
	var sentinelUUID uuid.UUID
	var owner, filenameHash string

	// get updated user
	userdata, err = GetUser(userdata.Username, userdata.password)
	if err != nil {
		userlib.DebugMsg("Could not update User struct!")
		return err
	}

	// verify DS
	senderDSPub, ok = userlib.KeystoreGet(sender + "-DSVerifyKey")
	if !ok {
		userlib.DebugMsg("Error retrieving sender DSVerifyKey")
		return errors.New(strings.ToTitle("Could not retrieve sender DSVerify Key!"))
	}
	magic_stringByte = []byte(magic_string)
	if len(magic_stringByte) < userlib.RSAKeySize/8 {
		userlib.DebugMsg("issue with retrieving magic string, incorrect array len")
		return errors.New(strings.ToTitle("magic_string is not correct length"))
	}
	signatureByte = magic_stringByte[:userlib.RSAKeySize/8]
	magicStringByte = magic_stringByte[userlib.RSAKeySize/8:]
	err = userlib.DSVerify(senderDSPub, magicStringByte, signatureByte)
	if err != nil {
		userlib.DebugMsg("Error verifiying DS in RecieveFile")
		return errors.New(strings.ToTitle("Could not verify sender signature!"))
	}

	// get symenc key and magicStringStruct containing file sentinel and owner name
	keyEnc = magicStringByte[:userlib.RSAKeySize/8]
	key, err = userlib.PKEDec(userdata.PKEDec, keyEnc)
	if err != nil {
		userlib.DebugMsg("Error decrypting symenc key in recievefile")
		return errors.New(strings.ToTitle("Could not retrieve SymEnc key from magic string!"))
	}
	if len(magicStringByte[256:])%userlib.AESBlockSize != 0 {
		userlib.DebugMsg("ciphertext is not a multiple of the block size!")
		return errors.New(strings.ToTitle("ciphertext is not a multiple of the block size!"))
	}
	plaintext = userlib.SymDec(key, magicStringByte[userlib.RSAKeySize/8:])
	plaintext = unpad(plaintext)
	err = json.Unmarshal(plaintext, &magicStringStruc)
	if err != nil {
		userlib.DebugMsg("Error unmarshalling magicStringStruc in recieveFile")
		return errors.New(strings.ToTitle("Could not unmarshall magicStringStruc!"))
	}
	sentinelUUID = magicStringStruc.SentinelUUID
	owner = magicStringStruc.Owner

	// verify filename does not exist in user's UUIDMap as a key
	filenameHash = getFilenameHash(filename, userdata.Username)
	if _, ok := userdata.UUIDMap[filenameHash]; ok {
		userlib.DebugMsg("File exists in user struct")
		return errors.New(strings.ToTitle("There is already a file with this name!"))
	}

	// verify owner is a valid user in Keystore
	_, ok = userlib.KeystoreGet(owner + "-PKEEncKey")
	if !ok {
		userlib.DebugMsg(owner + " does not exist! Could not find associated PKE Encryption Key")
		return errors.New(strings.ToTitle("Owner does not exist!"))
	}

	// modify userdata and save
	userdata.UUIDMap[filenameHash] = sentinelUUID
	userdata.OwnerMap[filenameHash] = owner
	err = postUser(*userdata, userdata.Username, userdata.password)
	if err != nil {
		userlib.DebugMsg("error posting user struct")
		return err
	}
	return nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	// variable declarations
	var filenameHash string
	var metadata Metadata
	var ok bool
	var newMetadataKey, newTextHMAC, newTextSymEnc, authStore, ciphertext, plaintext, authCompute, val []byte
	var text Text
	var sentinelUUID uuid.UUID
	var sentinel Sentinel
	var nodePKEEnc userlib.PKEEncKey

	userdata, err = GetUser(userdata.Username, userdata.password)
	if err != nil {
		userlib.DebugMsg("Could not update User struct!")
		return err
	}
	// verify user has access to file and is owner
	filenameHash = getFilenameHash(filename, userdata.Username)
	metadata, _, err = verifyFileAccess(*userdata, filenameHash)
	if err != nil {
		userlib.DebugMsg("could not verify that " + userdata.Username + " has access to " + filename)
		return errors.New(strings.ToTitle("could not verify that user has access to file"))
	}
	if metadata.Owner != userdata.Username {
		userlib.DebugMsg("User is not owner of the file")
		return errors.New(strings.ToTitle("User is not owner of file!"))
	}

	// verify target is valid and remove
	ok = accessContains(metadata.AccessMap, target_username)
	if !ok {
		userlib.DebugMsg("Target not in metadata accessList")
		return errors.New(strings.ToTitle("Target does not have access to the file!"))
	}
	ok = shared(userdata.Username, target_username, metadata.AccessMap)
	if !ok {
		userlib.DebugMsg("Owner did not share to target")
		return errors.New(strings.ToTitle("Owner did not share to target!"))
	}

	metadata.AccessMap, err = removeAccess(metadata.AccessMap, userdata.Username, target_username)
	if err != nil {
		return err
	}

	// recompute metadata key, textblock key, textblock HMAC key
	newMetadataKey = userlib.RandomBytes(userlib.AESBlockSize)
	newTextSymEnc = userlib.RandomBytes(userlib.AESBlockSize)
	newTextHMAC = userlib.RandomBytes(userlib.AESBlockSize)

	// reencrypt textblocks
	for i := range metadata.TextList {
		// retrieve from datastore and split val into components
		val, ok = userlib.DatastoreGet(metadata.TextList[i])
		if !ok {
			userlib.DebugMsg("text block is not at recorded UUID")
			return errors.New(strings.ToTitle("text block is not at recorded UUID"))
		}
		if len(val) < userlib.HashSize {
			userlib.DebugMsg("text block value was corrupted in datastore (not enough info)")
			return errors.New(strings.ToTitle("text block value was corrupted in datastore (not enough info)"))
		}
		authStore = val[:userlib.HashSize]
		ciphertext = val[userlib.HashSize:]

		// verify, decrypt, and unmarshal text block
		authCompute, err = userlib.HMACEval(metadata.TextMACKey, ciphertext)
		if err != nil {
			userlib.DebugMsg("error computing HMAC of ciphertext for text block")
			return err
		}
		if !userlib.HMACEqual(authStore, authCompute) {
			userlib.DebugMsg("cannot verify text block for " + filename)
			return errors.New(strings.ToTitle("cannot verify text block"))
		}
		if len(ciphertext)%userlib.AESBlockSize != 0 {
			userlib.DebugMsg("ciphertext is not a multiple of the block size!")
			return errors.New(strings.ToTitle("ciphertext is not a multiple of the block size!"))
		}
		plaintext = userlib.SymDec(metadata.TextEncKey, ciphertext)
		plaintext = unpad(plaintext)
		err = json.Unmarshal(plaintext, &text)
		if err != nil {
			userlib.DebugMsg("error unmarshalling text block")
			return err
		}
		if metadata.TextList[i] != text.TextUUID {
			userlib.DebugMsg("malicious user swapped text blocks!")
			return errors.New(strings.ToTitle("malicious user swapped text blocks!"))
		}

		// post text block
		err = postStruct(text, text.TextUUID, newTextSymEnc, newTextHMAC, nil, *userdata, false)
		if err != nil {
			userlib.DebugMsg("Error reposting text blocks")
			return err
		}
	}

	// Update file metadata and post
	metadata.TextEncKey = newTextSymEnc
	metadata.TextMACKey = newTextHMAC
	metadata.LastModified = userdata.Username
	err = postMetadata(*userdata, metadata, newMetadataKey)
	if err != nil {
		userlib.DebugMsg("error posting file metadata")
		return err
	}

	// Update file sentinel and post
	filenameHash = getFilenameHash(filename, userdata.Username)
	sentinelUUID, ok = userdata.UUIDMap[filenameHash]
	if !ok {
		userlib.DebugMsg("cannot find UUID of file sentinel!")
		return errors.New(strings.ToTitle("user cannot find UUID of file sentinel!"))
	}
	plaintext, ok = userlib.DatastoreGet(sentinelUUID)
	if !ok {
		userlib.DebugMsg("file sentinel is not at recorded UUID")
		return errors.New(strings.ToTitle("file sentinel is not at recorded UUID"))
	}
	err = json.Unmarshal(plaintext, &sentinel)
	if err != nil {
		userlib.DebugMsg("error unmarshalling file sentinel")
		return err
	}
	sentinel.MetadataKeyMap = make(map[string][]byte)
	for k := range metadata.AccessMap {
		nodePKEEnc, ok = userlib.KeystoreGet(k + "-PKEEncKey")
		if !ok {
			userlib.DebugMsg("Error finding node's PKEEnc Key")
			return errors.New(strings.ToTitle("Could not compute new PKEs for file sentinel."))
		}
		sentinel.MetadataKeyMap[k], err = userlib.PKEEnc(nodePKEEnc, newMetadataKey)
		if err != nil {
			userlib.DebugMsg("error RSA-encrypting file metadata key")
			return err
		}
	}
	val = getSentinelLockVal(metadata.MetadataUUID, newMetadataKey)
	sentinel.Lock, err = userlib.DSSign(userdata.DSSign, val)
	if err != nil {
		userlib.DebugMsg("error computing sentinel lock value")
		return err
	}

	// post file sentinel to userdata
	plaintext, err = json.Marshal(sentinel)
	if err != nil {
		userlib.DebugMsg("error marshalling file sentinel")
		return err
	}
	userlib.DatastoreSet(sentinelUUID, plaintext)
	return nil
}
