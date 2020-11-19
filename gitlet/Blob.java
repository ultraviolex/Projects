package gitlet;

import java.io.File;
import java.io.Serializable;

/** Blob class of gitlet.
 * @author Xiuhui Ming*/
public class Blob implements Serializable {

    /** Make a blob from a file.
     * @param f file. */
    public Blob(File f) {
        _fileName = f.getName();
        _contents = Utils.readContentsAsString(f);
        _ID = Utils.sha1("file", _contents);
    }

    /** Return filename.*/
    public String getfileName() {
        return _fileName;
    }

    /** Return contents. */
    public String getcontents() {
        return _contents;
    }

    /** Return ID. */
    public String getID() {
        return _ID;
    }

    /** File name. */
    private String _fileName;

    /** Contents of file. */
    private String _contents;

    /** SHA ID. */
    private String _ID;

}
