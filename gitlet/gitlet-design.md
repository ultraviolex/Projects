# Gitlet Design Document

**Name**: Xiuhui Ming

## Classes and Data Structures


###Commit
Stores log message and commit date, references to parental commits,
hashmap of file names to blobs

**Fields**

1. String log message: Message in string format
2. Date commit date: Date of commit
3. String parents: String shaID references to parent commit(s)
4. Hashmap files: Mapping of file names to blob references
5. SHA ID: Unique shaID that serves as reference 
to this commit
   
###Blob
Content of a file

**Fields**

1. String ID: Unique shaID that serves are reference to this blob
2. String file name: name of the file
3. String contents: contents of the file at commit time



## Algorithms
###Main

All algorithms check if there is an existing repository and
if the number of arguments is correct

**init**

If no existing repository, it initializes .gitlet directory in CWD,
staging area directory within .gitlet, as well as makes the 
first commit

**add**

Checks if file exists, then calls on commit add method

**commit**

Checks for commit message then checks if staging area directory is 
not empty, then makes a blob for each file in the staging area and
adds that to the _files hashmap of filename to blob ID. Updates
the shaID, parent, message, and time fields and then saves current 
commit to a file named shaID

**remove**

Checks if file is in the current working directory and if file is 
being tracked. If file exists in staging area, it removes the file. 
If file is being tracked, it removes the file from being tracked. 
If file is also in the CWD, it removes the file. Does not remove file
from CWD if file is not being tracked by current commit. 

**log**

Prints out logs of the commit starting with the most recent and ending 
with the initial commit by calling on commit.log()

**checkout**

General method for checkout, takes in a file name as a parameter.
If file exists in the current commit, then the file version in the 
current commit is added to the CWD, overwriting if needed. 

**checkout_File**

If args[] supplies only a file name for checkout, then method
calls on checkout passing in file name as args. 

**checkout_CommitFile**

If args[] supplies a commit ID and a file name, then current is 
set to the commit that is under commitID, if it is exists, and then
calls on checkout, passing in file name as args. 

**mergeCommit**

Specific method for merge commit, adds parent 2 and sets message.

**globalLog**

Print out log of all commits. 

**checkoutBranch**

Checkout the branchead of a given branch. 


**Find**

Returns the ID of all commits that have the given commit message. 

**Status**

Displays the status of gitlet, shows working branch and all branches, and
files staged for removal and addition. 


**Branch**

Creates a new branch with given name, the branchead is copy of the
headcommit. 


**rmBranch**

Removes a branch head. 


**Reset**

Resets gitlet to it's state at the given commit ID. 


**mergeCheck**

Merge precheck, checks for args and any failure cases. 


**Merge**

Merges the current and a given branch together. 


**MergeHelper**

Main method for merging a given file if needed. Checks to see
if merge is needed or not/what to do based off of the file. 
Returns a boolean (true if merged and false if not merged).


**Merge true**

Provides the contents of the files and calls on the actual 
file merging method.


**Checkoutstage**

Checkouts a file and stages it for addition. 


**MergeFiles**

Overwrites file contents with merged content. 


**Splitpoint**

Finds the latest common ancestor of two given branches. 


**LexicoSort**

Sorts a stringArray lexicographically. 

**Checkrepo**

Checks if repo initialized. 


**CheckargNum**

Checks if argument number is correct. 


**ExitWithError**

Exit and print out a given error message. 


###Commit
**add**

Adds file to staging area. If file is already being tracked, checks
to see if file being added has changes, if not, it does not add the file
and if that file exists in the staging area, it removes it from staging
area.

**log**

Prints out log of commit.


## Persistence
In the .gitlet directory, I have a folder BLOBS for blobs (stored as 
blobs written to files, name of each file is the blob ID), a folder COMMITS
for commits (same concept as blobs), a folder BRANCHES for branches (contains
branch head commit written to file, name of file is branch name), and a 
file HEAD that is my working head commit, which will always be the same as
one of the branch heads. There is also a staging area as well as a removal stage.
After each commit, I write to the COMMITS, BRANCHES folder and HEAD file
with the new commit, the BLOBS folder if there are blobs to add, and clear
the stages (removal and addition).  

