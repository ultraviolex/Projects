package gitlet;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Formatter;
import java.util.HashSet;

/** Gitlet application.
 * @author Xiuhui Ming
 */
public class Main {
    /** Current working directory. */
    static final File CWD = new File(".");

    /** Main metadata folder that holds old
     * copies of files and other metadata. */
    static final File GITLET = Utils.join(CWD, ".gitlet");

    /** Staging area. */
    static final File STAGING_AREA = Utils.join(GITLET, "stage");

    /** Blobs directory. */
    static final File BLOBS = Utils.join(GITLET, "blobs");

    /** Commits directory. */
    static final File COMMITS = Utils.join(GITLET, "commits");

    /** Branches directory. */
    static final File BRANCHES = Utils.join(GITLET, "branches");

    /** Stage for removal. */
    static final File RM_STAGE = Utils.join(GITLET, "removal_stage");

    /** Current head file path. */
    private static File headFile = Utils.join(GITLET, "head");

    /** Current commit. */
    private static Commit head;

    /**Current branch. */
    private static File branchHead;


    /** Usage: java gitlet.Main ARGS, where ARGS contains
     *  <COMMAND> <OPERAND> .... */
    public static void main(String[] args) {
        if (args.length == 0) {
            exitWithError("Please enter a command.");
        }
        switch (args[0]) {
        case "init":
            init(args);
            break;
        case "add":
            add(args);
            break;
        case "commit":
            commit(args);
            break;
        case "rm":
            remove(args);
            break;
        case "log":
            log(args);
            break;
        case "checkout":
            if (args.length == 3) {
                checkoutFile(args);
                break;
            } else if (args.length == 4) {
                checkoutCommitFile(args);
                break;
            } else if (args.length == 2) {
                checkoutBranch(args);
                break;
            } else {
                checkArgNum(args, 3);
                break;
            }
        case "global-log":
            globalLog(args);
            break;
        case "find":
            find(args);
            break;
        case "status":
            status(args);
            break;
        case "branch":
            branch(args);
            break;
        case "rm-branch":
            rmBranch(args);
            break;
        case "reset":
            reset(args);
            break;
        case "merge":
            merge(args);
            break;
        default:
            exitWithError("No command with that name exists.");
        }
    }

    /** Initialize a gitlet repo in CWD.
     * @param args */
    public static void init(String[] args) {
        checkArgNum(args, 1);
        if (GITLET.exists()) {
            exitWithError("A Gitlet version-control system "
                    + "already exists in the current directory.");
        } else {
            GITLET.mkdir();
            STAGING_AREA.mkdir();
            BLOBS.mkdir();
            COMMITS.mkdir();
            BRANCHES.mkdir();
            RM_STAGE.mkdir();
            File masterHead = Utils.join(BRANCHES, "master");
            head = new Commit();
            head.setMessage("initial commit");
            head.setTime("Wed Dec 31 16:00:00 1969 -0800");
            head.setParent1(null);
            head.setBranch("master");
            byte[] contents = Utils.serialize(head);
            head.setShaID(Utils.sha1("commit", contents));
            File initialCommit = Utils.join(COMMITS, head.getID());
            try {
                initialCommit.createNewFile();
                masterHead.createNewFile();
            } catch (IOException excp) {
                System.out.println("Error in main init.");
            }
            Utils.writeObject(masterHead, head);
            Utils.writeObject(initialCommit, head);
            try {
                headFile.createNewFile();
            } catch (IOException excp) {
                System.out.println("Error in main init.");
            }
            Utils.writeObject(headFile, head);
        }
    }

    /** Add file to staging area.
     * @param args  */
    public static void add(String[] args) {
        checkArgNum(args, 2);
        checkRepo();
        File add = Utils.join(CWD, args[1]);
        if (!add.exists()) {
            exitWithError("File does not exist.");
        }
        head = Utils.readObject(headFile, Commit.class);
        head.add(add);
    }

    /** Commit.
     * @param args */
    public static void commit(String[] args) {
        checkRepo();
        if (args.length == 1) {
            exitWithError("Please enter a commit message.");
        }
        checkArgNum(args, 2);
        if (args[1] == null || args[1].equals("")) {
            exitWithError("Please enter a commit message.");
        }
        head = Utils.readObject(headFile, Commit.class);
        branchHead = Utils.join(BRANCHES, head.getBranch());
        head.setMessage(args[1]);
        File[] staged = STAGING_AREA.listFiles();
        File[] rmStaged = RM_STAGE.listFiles();
        if (staged == null || staged.length == 0) {
            if (rmStaged == null || rmStaged.length == 0) {
                exitWithError("No changes added to the commit.");
            }
        }
        for (File f : staged) {
            Blob fBlob = new Blob(f);
            head.getFiles().put(f.getName(), fBlob.getID());
            File blob = Utils.join(BLOBS, fBlob.getID());
            try {
                blob.createNewFile();
            } catch (IOException excp) {
                System.out.println("Error in commit.");
            }
            Utils.writeObject(blob, fBlob);
            f.delete();
        }
        if (rmStaged.length != 0 && rmStaged != null) {
            for (File f : rmStaged) {
                if (head.getFiles().containsKey(f.getName())) {
                    head.getFiles().remove(f.getName());
                }
            }
            for (File f : rmStaged) {
                f.delete();
            }
        }
        Calendar cal = Calendar.getInstance();
        Formatter fmt = new Formatter();
        head.setTime(String.valueOf(fmt.format("%ta %tb %te %tT %tY %tz",
                cal, cal, cal, cal, cal, cal)));
        head.setParent1(head.getID());
        head.setParent2(null);
        head.getParents().add(head.getID());
        head.setShaID(Utils.sha1("commit", Utils.serialize(head)));
        Utils.writeObject(headFile, head);
        Utils.writeObject(branchHead, head);
        File commitFile = Utils.join(COMMITS, head.getID());
        try {
            commitFile.createNewFile();
        } catch (IOException excp) {
            exitWithError("Error in main commit.");
        }
        Utils.writeObject(commitFile, head);
    }

    /** Merge Commit.
     * @param args */
    public static void mergeCommit(String[] args) {
        head = Utils.readObject(headFile, Commit.class);
        branchHead = Utils.join(BRANCHES, head.getBranch());
        head.setMessage(args[1]);
        File[] staged = STAGING_AREA.listFiles();
        File[] rmStaged = RM_STAGE.listFiles();
        for (File f : staged) {
            Blob fBlob = new Blob(f);
            head.getFiles().put(f.getName(), fBlob.getID());
            File blob = Utils.join(BLOBS, fBlob.getID());
            try {
                blob.createNewFile();
            } catch (IOException excp) {
                System.out.println("Error in commit.");
            }
            Utils.writeObject(blob, fBlob);
        }
        for (File f : staged) {
            f.delete();
        }
        if (rmStaged.length != 0 && rmStaged != null) {
            for (File f : rmStaged) {
                if (head.getFiles().containsKey(f.getName())) {
                    head.getFiles().remove(f.getName());
                }
            }
            for (File f : rmStaged) {
                f.delete();
            }
        }
        Calendar cal = Calendar.getInstance();
        Formatter fmt = new Formatter();
        head.setTime(String.valueOf(fmt.format("%ta %tb %te %tT %tY %tz",
                cal, cal, cal, cal, cal, cal)));
        head.setParent1(head.getID());
        head.setParent2(args[2]);
        head.getParents().add(head.getID());
        head.getParents().add(args[2]);
        head.setShaID(Utils.sha1("commit", Utils.serialize(head)));
        Utils.writeObject(headFile, head);
        Utils.writeObject(branchHead, head);
        File commitFile = Utils.join(COMMITS, head.getID());
        try {
            commitFile.createNewFile();
        } catch (IOException excp) {
            exitWithError("Error in main commit.");
        }
        Utils.writeObject(commitFile, head);
    }

    /** Remove a file.
     * @param args */
    public static void remove(String[] args) {
        checkRepo();
        checkArgNum(args, 2);
        head = Utils.readObject(headFile, Commit.class);
        File fStage = Utils.join(STAGING_AREA, args[1]);
        File fCWD = Utils.join(CWD, args[1]);
        if (!fStage.exists() && !head.getFiles().containsKey(args[1])) {
            exitWithError("No reason to remove the file.");
        }
        String fContents;
        if (fStage.exists()) {
            fContents = Utils.readContentsAsString(fStage);
            fStage.delete();
        } else {
            Blob fBlob = Utils.readObject(Utils.join(BLOBS,
                    head.getFiles().get(args[1])), Blob.class);
            fContents = fBlob.getcontents();
        }
        if (head.getFiles().containsKey(args[1])) {
            File fRmStage = Utils.join(RM_STAGE, args[1]);
            try {
                fRmStage.createNewFile();
                Utils.writeContents(fRmStage, fContents);
            } catch (IOException excp) {
                System.out.println("Error in rm main.");
            }
            if (fCWD.exists()) {
                fCWD.delete();
            }
        }
    }

    /** Display log.
     * @param args */
    public static void log(String[] args) {
        checkRepo();
        checkArgNum(args, 1);
        head = Utils.readObject(headFile, Commit.class);
        while (head.getParent1() != null) {
            head.log();
            String parentID = head.getParent1();
            File parentCommitFile = Utils.join(COMMITS, parentID);
            head = Utils.readObject(parentCommitFile, Commit.class);
        }
        System.out.println("===");
        System.out.println("commit " + head.getID());
        if (head.getParent2() != null) {
            System.out.println("Merge: " + head.getParent1().substring(0, 7)
                    + " " + head.getParent2().substring(0, 7));
        }
        System.out.println("Date: " + head.getTime());
        System.out.println(head.getMessage());
        System.out.println();
    }

    /** Print out global log.
     * @param args */
    public static void globalLog(String[] args) {
        checkRepo();
        checkArgNum(args, 1);
        for (File f : COMMITS.listFiles()) {
            Commit c = Utils.readObject(f, Commit.class);
            c.log();
        }
    }

    /** Checkout.
     * @param args */
    public static void checkoutFile(String[] args) {
        checkRepo();
        if (!args[1].equals("--")) {
            exitWithError("Incorrect operands.");
        }
        head = Utils.readObject(headFile, Commit.class);
        checkout(args[2]);
    }

    /**Checkout general.
     * @param file */
    private static void checkout(String file) {
        if (!head.getFiles().containsKey(file)) {
            exitWithError("File does not exist in that commit.");
        }
        File checkoutFile = Utils.join(BLOBS, head.getFiles().get(file));
        Blob checkoutBlob = Utils.readObject(checkoutFile, Blob.class);
        File overwrite = Utils.join(CWD, file);
        if (!overwrite.exists()) {
            try {
                overwrite.createNewFile();
            } catch (IOException excp) {
                System.out.println("Error in Main checkout_File");
            }
        }
        Utils.writeContents(overwrite, checkoutBlob.getcontents());
    }

    /** Checkout given ID and file name.
     * @param args */
    public static void checkoutCommitFile(String[] args) {
        checkRepo();
        if (!args[2].equals("--")) {
            exitWithError("Incorrect operands.");
        }
        String fName = args[1];
        if (args[1].length() < SHAID_LENGTH) {
            ArrayList<String> names = new ArrayList<>();
            for (String f : COMMITS.list()) {
                if (f.startsWith(fName)) {
                    names.add(f);
                }
            }
            if (names.size() != 1) {
                exitWithError("No commit with that ID exists.");
            } else {
                fName = names.get(0);
            }
        }
        File commitFile = Utils.join(COMMITS, fName);
        if (!commitFile.exists()) {
            exitWithError("No commit with that ID exists.");
        }
        head = Utils.readObject(commitFile, Commit.class);
        checkout(args[3]);
    }

    /** Checkout given branch.
     * @param args */
    public static void checkoutBranch(String[] args) {
        checkRepo();
        branchHead = Utils.join(BRANCHES, args[1]);
        if (!branchHead.exists()) {
            exitWithError("No such branch exists.");
        }
        head = Utils.readObject(headFile, Commit.class);
        if (args[1].equals(head.getBranch())) {
            exitWithError("No need to checkout the current branch.");
        }
        Commit branchHeadCommit = Utils.readObject(branchHead, Commit.class);
        for (String fName : branchHeadCommit.getFiles().keySet()) {
            if (Utils.join(CWD, fName).exists()
                    && !head.getFiles().containsKey(fName)) {
                exitWithError("There is an untracked file in the way; "
                        + "delete it, or add and commit it first.");
            }
        }
        for (String fName : head.getFiles().keySet()) {
            if (Utils.join(CWD, fName).exists()
                    && !branchHeadCommit.getFiles().containsKey(fName)) {
                Utils.join(CWD, fName).delete();
            }
        }
        for (String fName : branchHeadCommit.getFiles().keySet()) {
            File fBlobFile = Utils.join(BLOBS,
                    branchHeadCommit.getFiles().get(fName));
            Blob fBlob = Utils.readObject(fBlobFile, Blob.class);
            File fCWD = Utils.join(CWD, fName);
            if (!fCWD.exists()) {
                try {
                    fCWD.createNewFile();
                } catch (IOException excp) {
                    System.out.println("Error in Main brach Checkout.");
                }
            }
            Utils.writeContents(fCWD, fBlob.getcontents());
        }
        if (args[1] != head.getBranch()) {
            File[] staged = STAGING_AREA.listFiles();
            if (staged != null && staged.length != 0) {
                for (File f : staged) {
                    f.delete();
                }
            }
        }
        Utils.writeObject(headFile, branchHeadCommit);
    }


    /** Find IDs of all commits that have given commit message.
     * @param args */
    public static void find(String[] args) {
        checkRepo();
        checkArgNum(args, 2);
        ArrayList<Commit> matches = new ArrayList<>();
        for (File f : COMMITS.listFiles()) {
            Commit fCommit = Utils.readObject(f, Commit.class);
            if (fCommit.getMessage().equals(args[1])) {
                matches.add(fCommit);
            }
        }
        if (matches.size() == 0) {
            exitWithError("Found no commit with that message.");
        }
        for (Commit c : matches) {
            System.out.println(c.getID());
        }
    }

    /** Displays status of .gitlet.
     * @param args */
    public static void status(String[] args) {
        checkRepo();
        checkArgNum(args, 1);
        head = Utils.readObject(headFile, Commit.class);
        String[] branches = BRANCHES.list();
        branches = lexicoSort(branches);
        System.out.println("=== Branches ===");
        for (String s : branches) {
            if (s.equals(head.getBranch())) {
                System.out.println("*" + s);
            } else {
                System.out.println(s);
            }
        }
        System.out.println();
        String[] staged = STAGING_AREA.list();
        System.out.println("=== Staged Files ===");
        if (staged.length != 0) {
            staged = lexicoSort(staged);
            for (String s : staged) {
                System.out.println(s);
            }
        }
        System.out.println();
        System.out.println("=== Removed Files ===");
        String[] rmStaged = RM_STAGE.list();
        if (rmStaged.length != 0) {
            rmStaged = lexicoSort(rmStaged);
            for (String s: rmStaged) {
                System.out.println(s);
            }
        }
        System.out.println();
        System.out.println("=== Modifications Not Staged For Commit ===");
        System.out.println();
        System.out.println("=== Untracked Files ===");
        System.out.println();
    }

    /** Branch method.
     * @param args */
    public static void branch(String[] args) {
        checkRepo();
        checkArgNum(args, 2);
        head = Utils.readObject(headFile, Commit.class);
        File newBranch = Utils.join(BRANCHES, args[1]);
        if (newBranch.exists()) {
            exitWithError("A branch with that name already exists.");
        }
        try {
            newBranch.createNewFile();
        } catch (IOException excp) {
            System.out.println("Error in main branch.");
        }
        head.setBranch(args[1]);
        Utils.writeObject(newBranch, head);
    }

    /** Remove branch.
     * @param args */
    public static void rmBranch(String[] args) {
        checkRepo();
        checkArgNum(args, 2);
        head = Utils.readObject(headFile, Commit.class);
        File rmBranch = Utils.join(BRANCHES, args[1]);
        if (!rmBranch.exists()) {
            exitWithError("A branch with that name does not exist.");
        }
        if (head.getBranch().equals(args[1])) {
            exitWithError("Cannot remove the current branch.");
        }
        rmBranch.delete();
    }

    /** Reset.
     * @param args */
    public static void reset(String[] args) {
        checkRepo();
        checkArgNum(args, 2);
        head = Utils.readObject(headFile, Commit.class);
        File commitFile = Utils.join(COMMITS, args[1]);
        if (!commitFile.exists()) {
            exitWithError("No commit with that id exists.");
        }
        Commit c = Utils.readObject(commitFile, Commit.class);
        for (String fName : c.getFiles().keySet()) {
            if (Utils.join(CWD, fName).exists()
                    && !head.getFiles().containsKey(fName)) {
                exitWithError("There is an untracked file in the way; "
                        + "delete it, or add and commit it first.");
            }
        }
        head = c;
        for (String fName : c.getFiles().keySet()) {
            checkout(fName);
        }
        head = Utils.readObject(headFile, Commit.class);
        for (String fName : head.getFiles().keySet()) {
            if (!c.getFiles().containsKey(fName)) {
                File fCWD = Utils.join(CWD, fName);
                if (fCWD.exists()) {
                    fCWD.delete();
                }
            }
        }
        File currentBranchHead = Utils.join(BRANCHES, head.getBranch());
        Utils.writeObject(currentBranchHead, c);
        Utils.writeObject(headFile, c);
        File[] staged = STAGING_AREA.listFiles();
        if (staged != null && staged.length != 0) {
            for (File f : staged) {
                f.delete();
            }
        }
    }

    /** Merge precheck.
     * ARGS
     */
    public static void mergeCheck(String[] args) {
        checkRepo();
        checkArgNum(args, 2);
        File givenBranch = Utils.join(BRANCHES, args[1]);
        head = Utils.readObject(headFile, Commit.class);
        if (!givenBranch.exists()) {
            exitWithError("A branch with that name does not exist.");
        }
        Commit givenCommit = Utils.readObject(givenBranch, Commit.class);
        for (String fName : givenCommit.getFiles().keySet()) {
            if (Utils.join(CWD, fName).exists()
                    && !Utils.join(STAGING_AREA, fName).exists()) {
                if (!head.getFiles().containsKey(fName)) {
                    exitWithError("There is an untracked file in the way; "
                            + "delete it, or add and commit it first.");
                }
            }
        }
        if (head.getBranch().equals(givenCommit.getBranch())) {
            exitWithError("Cannot merge a branch with itself.");
        }
        File[] staged = STAGING_AREA.listFiles();
        File[] removing = RM_STAGE.listFiles();
        if (staged.length != 0) {
            exitWithError("You have uncommitted changes.");
        }
        if (removing.length != 0) {
            exitWithError("You have uncommitted changes.");
        }
    }

    /** Merge.
     * @param args */
    public static void merge(String[] args) {
        mergeCheck(args);
        Commit givenCommit = Utils.readObject
                (Utils.join(BRANCHES, args[1]), Commit.class);
        head = Utils.readObject(headFile, Commit.class);
        ArrayList<Boolean> mergeConflict = new ArrayList<>();
        String[] splitPointArray = splitPoint(head, givenCommit, 0);
        Commit splitPoint = Utils.readObject(Utils.join(COMMITS,
                splitPointArray[0]), Commit.class);
        if (splitPoint.getID().equals(givenCommit.getID())) {
            System.out.println("Given branch is an "
                    + "ancestor of the current branch.");
            return;
        }
        if (splitPoint.getID().equals(head.getID())) {
            checkoutBranch(args);
            System.out.println("Current branch fast-forwarded.");
            return;
        }
        HashSet<String> allFiles = new HashSet<>();
        allFiles.addAll(splitPoint.getFiles().keySet());
        allFiles.addAll(givenCommit.getFiles().keySet());
        allFiles.addAll(head.getFiles().keySet());
        for (String fName : allFiles) {
            mergeConflict.add(mergeHelper(fName,
                    givenCommit, head, splitPoint));
        }
        if (splitPoint != head && splitPoint != givenCommit) {
            String[] cArgs = {"commit", "Merged " + givenCommit.getBranch()
                    + " into " + head.getBranch() + ".", givenCommit.getID()};
            mergeCommit(cArgs);
        }
        if (mergeConflict.contains(true)) {
            System.out.println("Encountered a merge conflict.");
        }
    }

    /** Merge helper.
     * FILES GIVEN CURRENT FNAME SPLITPOINT
     * @return merge conflict true or not
     */
    public static boolean mergeHelper(String fName, Commit given,
                                      Commit current, Commit splitPoint) {
        boolean mergeConflict = false;
        if (given.getFiles().containsKey(fName)
                && current.getFiles().containsKey(fName)) {
            File gBlobFile = Utils.join(BLOBS,
                    given.getFiles().get(fName));
            Blob gBlob = Utils.readObject(gBlobFile, Blob.class);
            File cBlobFile = Utils.join(BLOBS, current.getFiles().get(fName));
            Blob cBlob = Utils.readObject(cBlobFile, Blob.class);
            if (!gBlob.getcontents().equals(cBlob.getcontents())) {
                if (splitPoint.getFiles().containsKey(fName)) {
                    if (splitPoint.getFiles().get(fName).
                            equals(current.getFiles().get(fName))) {
                        checkoutStage(given, fName);
                    } else if (splitPoint.getFiles().get(fName).
                            equals(given.getFiles().get(fName))) {
                        return mergeConflict;
                    } else {
                        mergeConflict = true;
                        mergeTrue(given, current, fName);
                    }
                } else {
                    mergeConflict = true;
                    mergeTrue(given, current, fName);
                }
            }
        } else if (given.getFiles().containsKey(fName)) {
            if (!splitPoint.getFiles().containsKey(fName)) {
                checkoutStage(given, fName);
            } else {
                if (!given.getFiles().get(fName).
                        equals(splitPoint.getFiles().get(fName))) {
                    mergeConflict = true;
                    mergeTrue(given, current, fName);
                }
            }
        } else if (current.getFiles().containsKey(fName)) {
            if (splitPoint.getFiles().containsKey(fName)) {
                if (splitPoint.getFiles().get(fName).
                        equals(current.getFiles().get(fName))) {
                    String[] rmArgs = {"rm", fName};
                    remove(rmArgs);
                } else {
                    mergeConflict = true;
                    mergeTrue(given, current, fName);
                }
            }
        }
        return mergeConflict;
    }

    /**Merge method and commit.
     * GIVEN CURRENT NAME */
    public static void mergeTrue(Commit given, Commit current, String name) {
        if (given.getFiles().containsKey(name)
                && current.getFiles().containsKey(name)) {
            File gBlobFile = Utils.join(BLOBS, given.getFiles().get(name));
            Blob gBlob = Utils.readObject(gBlobFile, Blob.class);
            File cBlobFile = Utils.join(BLOBS, current.getFiles().get(name));
            Blob cBlob = Utils.readObject(cBlobFile, Blob.class);
            mergeFiles(gBlob.getcontents(), cBlob.getcontents(), name);
        } else if (given.getFiles().containsKey(name)) {
            File gBlobFile = Utils.join(BLOBS, given.getFiles().get(name));
            Blob gBlob = Utils.readObject(gBlobFile, Blob.class);
            mergeFiles(gBlob.getcontents(), "", name);
        } else {
            File cBlobFile = Utils.join(BLOBS, current.getFiles().get(name));
            Blob cBlob = Utils.readObject(cBlobFile, Blob.class);
            mergeFiles("", cBlob.getcontents(), name);
        }
    }

    /** Checkout and stage a file from a given commit.
     * C FNAME*/
    public static void checkoutStage(Commit c, String fName) {
        head = c;
        checkout(fName);
        String contents = Utils.readObject(Utils.join(BLOBS,
                c.getFiles().get(fName)), Blob.class).getcontents();
        head = Utils.readObject(headFile, Commit.class);
        File fStage = Utils.join(STAGING_AREA, fName);
        try {
            fStage.createNewFile();
        } catch (IOException excp) {
            System.out.println("error in checkoutstage.");
        }
        Utils.writeContents(fStage, contents);
    }

    /** Merge files.
     * GIVEN CURRENT NAME
     */
    public static void mergeFiles(String given, String current, String name) {
        File fCWD = Utils.join(CWD, name);
        if (!fCWD.exists()) {
            try {
                fCWD.createNewFile();
            } catch (IOException excp) {
                System.out.println("error in mergeFiles.");
            }
        }
        String contents = "<<<<<<< HEAD\n" + current
                + "=======\n" + given + ">>>>>>>\n";
        File fStage = Utils.join(STAGING_AREA, name);
        try {
            fStage.createNewFile();
        } catch (IOException excp) {
            System.out.println("error in checkoutstage.");
        }
        Utils.writeContents(fStage, contents);
        Utils.writeContents(fCWD, contents);
    }

    /** Find the split point.
     * CURRENT GIVEN STEPS
     * @return commit*/
    public static String[] splitPoint(Commit current, Commit given, int steps) {
        String[] option1 = new String[2];
        String[] option2 = new String[2];
        while (!given.getParents().contains(current.getID())
                && !given.getID().equals(current.getID())) {
            steps += 1;
            if (current.getParent2() != null) {
                option2 = splitPoint(
                        Utils.readObject(Utils.join
                                (COMMITS, current.getParent2()), Commit.class),
                        given, steps);
            }
            current = Utils.readObject(Utils.join
                    (COMMITS, current.getParent1()), Commit.class);
        }
        option1[0] = current.getID();
        option1[1] = Integer.toString(steps);
        if (option2[1] != null && Integer.parseInt(option2[1])
                < Integer.parseInt(option1[1])) {
            return option2;
        }
        return option1;
    }


    /** Sort in lexicographic order.
     * ARGS
     * @return sorted*/
    public static String[] lexicoSort(String[] args) {
        for (int i = 0; i < args.length - 1; i += 1) {
            for (int j = i + 1; j < args.length; j += 1) {
                if (args[1].compareTo(args[j]) > 0) {
                    String s = args[i];
                    args[i] = args[j];
                    args[j] = s;
                }
            }
        }
        return args;
    }


    /** Check if repo exists. */
    public static void checkRepo() {
        if (!GITLET.exists()) {
            exitWithError("Not in an initialized Gitlet directory.");
        }
    }

    /** Check number of args to be correct.
     * ARGS N */
    public static void checkArgNum(String[] args, int n) {
        if (args.length != n) {
            exitWithError("Incorrect operands.");
        }
    }

    /** Exit program with error message MESSAGE and exit code 0.
     * @param message */
    public static void exitWithError(String message) {
        if (message != null && !message.equals("")) {
            System.out.println(message);
        }
        System.exit(0);
    }

    /** ShaID Length. */
    private static final int SHAID_LENGTH = 40;
}
