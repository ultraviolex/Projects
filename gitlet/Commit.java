package gitlet;
import java.io.File;
import java.io.IOException;
import java.io.Serializable;
import java.util.HashMap;
import java.util.HashSet;


/** Commit object and it's methods.
 * @author Xiuhui Ming */
public class Commit implements Serializable {

    /** CWD. */
    static final File CWD = Main.CWD;
    /** GITLET. */
    static final File GITLET = Main.GITLET;
    /** Staging. */
    static final File STAGING_AREA = Main.STAGING_AREA;
    /** Stage for removal. */
    static final File RM_STAGE = Utils.join(GITLET, "removal_stage");
    /** Blobs. */
    static final File BLOBS = Utils.join(GITLET, "blobs");

    /** Create new commit. */
    public Commit() {
    }

    /** Print out log of commit. */
    public void log() {
        System.out.println("===");
        System.out.println("commit " + _shaID);
        if (parent2 != null) {
            System.out.println("Merge: " + parent1.substring(0, 7)
                    + " " + parent2.substring(0, 7));
        }
        System.out.println("Date: " + _time);
        System.out.println(_message);
        System.out.println();
    }

    /** Add a file to the stage.
     * @param f file */
    public void add(File f) {
        String fContents = Utils.readContentsAsString(f);
        File rmStage = Utils.join(RM_STAGE, f.getName());
        if (rmStage.exists()) {
            rmStage.delete();
        }
        if (_files.containsKey(f.getName())) {
            File prevFile = Utils.join(BLOBS, _files.get(f.getName()));
            String prevContents = Utils.readObject
                    (prevFile, Blob.class).getcontents();
            if (prevContents.equals(fContents)) {
                File staged = Utils.join(STAGING_AREA, f.getName());
                if (staged.exists()) {
                    staged.delete();
                }
                return;
            }
        }
        File stage = Utils.join(STAGING_AREA, f.getName());
        if (!stage.exists()) {
            try {
                stage.createNewFile();
            } catch (IOException e) {
                System.out.println("Error in Commit.add.");
            }
        }
        Utils.writeContents(stage, fContents);
    }

    /** Get parent1.
     * @return parent1 */
    public String getParent1() {
        return parent1;
    }

    /** Get parent2.
     * @return parent2 */
    public String getParent2() {
        return parent2;
    }

    /** Get message.
     * @return message */
    public String getMessage() {
        return _message;
    }

    /** Get time.
     * @return time */
    public String getTime() {
        return _time;
    }

    /** Get ID.
     * @return ID */
    public String getID() {
        return _shaID;
    }

    /** Get tracked files.
     * @return files */
    public HashMap<String, String> getFiles() {
        return _files;
    }

    /** Get ancestors hashmap.
     * @return parents */
    public HashSet<String> getParents() {
        return _parents;
    }

    /** Get branch.
     * @return branch */
    public String getBranch() {
        return _branch;
    }

    /** Set branch.
     * @param branch b
     * */
    public void setBranch(String branch) {
        _branch = branch;
    }

    /** Set parent1.
     * @param parent string */
    public void setParent1(String parent) {
        parent1 = parent;
    }

    /** Set parent2.
     * @param parent string */
    public void setParent2(String parent) {
        parent2 = parent;
    }

    /** Set message.
     * @param message m
     */
    public void setMessage(String message) {
        _message = message;
    }

    /** Set time.
     * @param time t
     */
    public void setTime(String time) {
        _time = time;
    }

    /** Set _shaID.
     * @param id string
     */
    public void setShaID(String id) {
        _shaID = id;
    }

    /** Commit parent 1. */
    private String parent1 = null;

    /** Commit parent 2. */
    private String parent2 = null;

    /** Commit message. */
    private String _message;

    /** ID of commit. */
    private String _shaID;

    /** Date of commit. */
    private String _time;

    /** Branch commit is in. */
    private String _branch;

    /** Hashmap of file name to blob references. */
    private HashMap<String, String> _files = new HashMap<String, String>();

    /** Hashset of all this commits ancestor commit IDS of this commit. */
    private HashSet<String> _parents = new HashSet<>();

}
