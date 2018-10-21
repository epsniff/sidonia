package cloudstorage

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/lytics/cloudstorage/logging"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const GCSFSStorageSource = "gcsFS"

var (
	GCSRetries int = 55

	// Ensure we implement ObjectIterator
	_ ObjectIterator = (*GcsObjectIterator)(nil)
)

// GcsFS Simple wrapper for accessing smaller GCS files, it doesn't currently implement a
// Reader/Writer interface so not useful for stream reading of large files yet.
type GcsFS struct {
	gcs       *storage.Client
	bucket    string
	cachepath string
	PageSize  int //TODO pipe this in from eventstore
	Id        string

	Log logging.Logger
}

func NewGCSStore(gcs *storage.Client, bucket, cachepath string, pagesize int, l logging.Logger) (*GcsFS, error) {
	err := os.MkdirAll(path.Dir(cachepath), 0775)
	if err != nil {
		return nil, fmt.Errorf("unable to create path. path=%s err=%v", cachepath, err)
	}

	uid := uuid.NewUUID().String()
	uid = strings.Replace(uid, "-", "", -1)

	return &GcsFS{
		gcs:       gcs,
		bucket:    bucket,
		cachepath: cachepath,
		Id:        uid,
		PageSize:  pagesize,
		Log:       l,
	}, nil
}

func (g *GcsFS) String() string {
	return fmt.Sprintf("gs://%s/", g.bucket)
}

func (g *GcsFS) gcsb() *storage.BucketHandle {
	return g.gcs.Bucket(g.bucket)
}

func (g *GcsFS) NewObject(objectname string) (Object, error) {
	obj, err := g.Get(objectname)
	if err != nil && err != ObjectNotFound {
		return nil, err
	} else if obj != nil {
		return nil, ObjectExists
	}

	cf := cachepathObj(g.cachepath, objectname, g.Id)

	return &gcsFSObject{
		name:       objectname,
		metadata:   map[string]string{ContextTypeKey: contentType(objectname)},
		gcsb:       g.gcsb(),
		bucket:     g.bucket,
		cachedcopy: nil,
		cachepath:  cf,
		log:        g.Log,
	}, nil
}

// Get Gets a single File Object
func (g *GcsFS) Get(objectpath string) (Object, error) {

	gobj, err := g.gcsb().Object(objectpath).Attrs(context.Background()) // .Objects(context.Background(), q)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return nil, ObjectNotFound
		}
		g.Log.Errorf("couldn't get object. obj=%q err=%v", objectpath, err)
		return nil, err
	}

	if gobj == nil {
		return nil, ObjectNotFound
	}

	return newObjectFromGcs(g, gobj), nil
}

func (g *GcsFS) List(query Query) (Objects, error) {

	var q = &storage.Query{Prefix: query.Prefix}

	res, err := g.listObjects(q, GCSRetries)
	if err != nil {
		g.Log.Warnf("couldn't list objects. prefix=%s err=%v", q.Prefix, err)
		return nil, err
	}

	if res == nil {
		return make(Objects, 0), nil
	}

	res = query.applyFilters(res)

	return res, nil
}

// Objects returns an iterator over the objects in the google bucket that match the Query q.
// If q is nil, no filtering is done.
func (g *GcsFS) Objects(ctx context.Context, csq Query) ObjectIterator {
	var q = &storage.Query{Prefix: csq.Prefix}
	iter := g.gcsb().Objects(ctx, q)
	return &GcsObjectIterator{g, ctx, iter}
}

// ListObjects iterates to find a list of objects
func (g *GcsFS) listObjects(q *storage.Query, retries int) (Objects, error) {
	var lasterr error = nil

	for i := 0; i < retries; i++ {
		objects := make(Objects, 0)
		iter := g.gcsb().Objects(context.Background(), q)
	iterLoop:
		for {
			oa, err := iter.Next()
			switch err {
			case nil:
				objects = append(objects, newObjectFromGcs(g, oa))
			case iterator.Done:
				return objects, nil
			default:
				g.Log.Warnf("error listing objects for the bucket. try:%d store:%s q.prefix:%v err:%v", i, g, q.Prefix, err)
				lasterr = err
				backoff(i)
				break iterLoop
			}
		}
	}
	return nil, lasterr
}

// Folders
func (g *GcsFS) Folders(ctx context.Context, csq Query) ([]string, error) {
	var q = &storage.Query{Delimiter: csq.Delimiter, Prefix: csq.Prefix}
	iter := g.gcsb().Objects(ctx, q)
	folders := make([]string, 0)
	for {
		select {
		case <-ctx.Done():
			// If has been closed
			return folders, ctx.Err()
		default:
			o, err := iter.Next()
			if err == nil {
				folders = append(folders, o.Prefix)
			} else if err == iterator.Done {
				return folders, nil
			} else if err == context.Canceled || err == context.DeadlineExceeded {
				// Return to user
				return nil, err
			}
		}
	}
	panic("unreacheable")
}

func (g *GcsFS) NewReader(o string) (io.ReadCloser, error) {
	return g.NewReaderWithContext(context.Background(), o)
}
func (g *GcsFS) NewReaderWithContext(ctx context.Context, o string) (io.ReadCloser, error) {
	rc, err := g.gcsb().Object(o).NewReader(ctx)
	if err == storage.ErrObjectNotExist {
		return rc, ObjectNotFound
	}
	return rc, err
}

func (g *GcsFS) NewWriter(o string, metadata map[string]string) (io.WriteCloser, error) {
	return g.NewWriterWithContext(context.Background(), o, metadata)
}
func (g *GcsFS) NewWriterWithContext(ctx context.Context, o string, metadata map[string]string) (io.WriteCloser, error) {
	wc := g.gcsb().Object(o).NewWriter(ctx)
	if metadata != nil {
		wc.Metadata = metadata
		//contenttype is only used for viewing the file in a browser. (i.e. the GCS Object browser).
		ctype := ensureContextType(o, metadata)
		wc.ContentType = ctype
	}
	return wc, nil
}

func (g *GcsFS) Delete(obj string) error {
	err := g.gcsb().Object(obj).Delete(context.Background())
	if err != nil {
		g.Log.Errorf("error deleting object. object=%s%s err=%v", g, obj, err)
		return err
	}
	return nil
}

type GcsObjectIterator struct {
	g    *GcsFS
	ctx  context.Context
	iter *storage.ObjectIterator
}

func (it *GcsObjectIterator) Next() (Object, error) {
	var lasterr error = nil
	retryCt := 0

	for {
		select {
		case <-it.ctx.Done():
			// If has been closed
			return nil, it.ctx.Err()
		default:
			o, err := it.iter.Next()
			if err == nil {
				return newObjectFromGcs(it.g, o), nil
			} else if err == iterator.Done {
				return nil, err
			} else if err == context.Canceled || err == context.DeadlineExceeded {
				// Return to user
				return nil, err
			}
			lasterr = err
			if retryCt < 5 {
				backoff(retryCt)
			} else {
				return nil, err
			}
			retryCt++
		}
	}

	return nil, lasterr
}

type gcsFSObject struct {
	name         string
	updated      time.Time
	metadata     map[string]string
	googleObject *storage.ObjectAttrs

	gcsb   *storage.BucketHandle
	bucket string

	cachedcopy *os.File
	readonly   bool
	opened     bool

	cachepath string
	log       logging.Logger
}

func newObjectFromGcs(g *GcsFS, o *storage.ObjectAttrs) *gcsFSObject {
	return &gcsFSObject{
		name:      o.Name,
		updated:   o.Updated,
		metadata:  o.Metadata,
		gcsb:      g.gcsb(),
		bucket:    g.bucket,
		cachepath: cachepathObj(g.cachepath, o.Name, g.Id),
		log:       g.Log,
	}
}
func (o *gcsFSObject) StorageSource() string {
	return GCSFSStorageSource
}
func (o *gcsFSObject) Name() string {
	return o.name
}
func (o *gcsFSObject) String() string {
	return o.name
}
func (o *gcsFSObject) Updated() time.Time {
	return o.updated
}
func (o *gcsFSObject) MetaData() map[string]string {
	return o.metadata
}
func (o *gcsFSObject) SetMetaData(meta map[string]string) {
	o.metadata = meta
}

func (o *gcsFSObject) Delete() error {
	o.Release()
	err := o.gcsb.Object(o.name).Delete(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (o *gcsFSObject) Open(accesslevel AccessLevel) (*os.File, error) {
	if o.opened {
		return nil, fmt.Errorf("the store object is already opened. %s", o.name)
	}

	var errs []error = make([]error, 0)
	var cachedcopy *os.File = nil
	var err error
	var readonly = accesslevel == ReadOnly

	err = os.MkdirAll(path.Dir(o.cachepath), 0775)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating cachedcopy dir. cachepath=%s object=%s err=%v",
			o.cachepath, o.name, err)
	}

	err = ensureDir(o.cachepath)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating cachedcopy's dir. cachepath=%s err=%v",
			o.cachepath, err)
	}

	cachedcopy, err = os.Create(o.cachepath)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating file. local=%s err=%v",
			o.cachepath, err)
	}

	for try := 0; try < GCSRetries; try++ {
		if o.googleObject == nil {
			gobj, err := o.gcsb.Object(o.name).Attrs(context.Background())
			if err != nil {
				if strings.Contains(err.Error(), "doesn't exist") {
					// New, this is fine
				} else {
					errs = append(errs, fmt.Errorf("error storage.NewReader err=%v", err))
					o.log.Debugf("error fetching object %q err=%v", o.name, errs)
					backoff(try)
					continue
				}
			}

			if gobj != nil {
				o.googleObject = gobj
			}
		}

		if o.googleObject != nil {
			//we have a preexisting object, so lets download it..
			rc, err := o.gcsb.Object(o.name).NewReader(context.Background())
			if err != nil {
				errs = append(errs, fmt.Errorf("error storage.NewReader err=%v", err))
				o.log.Debugf("%v", errs)
				backoff(try)
				continue
			}
			defer rc.Close()

			if _, err := cachedcopy.Seek(0, os.SEEK_SET); err != nil {
				return nil, fmt.Errorf("error seeking to start of cachedcopy err=%v", err) //don't retry on local fs errors
			}

			_, err = io.Copy(cachedcopy, rc)
			if err != nil {
				errs = append(errs, fmt.Errorf("error coping bytes. err=%v", err))
				o.log.Debugf("%v", errs)
				//recreate the cachedcopy file incase it has incomplete data
				if err := os.Remove(o.cachepath); err != nil {
					return nil, fmt.Errorf("error resetting the cachedcopy err=%v", err) //don't retry on local fs errors
				}
				if cachedcopy, err = os.Create(o.cachepath); err != nil {
					return nil, fmt.Errorf("error occurred creating a new cachedcopy file. local=%s err=%v",
						o.cachepath, err)
				}

				backoff(try)
				continue
			}
		}

		if readonly {
			cachedcopy.Close()
			cachedcopy, err = os.Open(o.cachepath)
			if err != nil {
				name := "unknown"
				if cachedcopy != nil {
					name = cachedcopy.Name()
				}
				return nil, fmt.Errorf("error occurred opening file. local=%s object=%s tfile=%v err=%v",
					o.cachepath, o.name, name, err)
			}
		}

		o.cachedcopy = cachedcopy
		o.readonly = readonly
		o.opened = true
		return o.cachedcopy, nil
	}

	return nil, fmt.Errorf("fetch error retry cnt reached: obj=%s tfile=%v errs:[%v]",
		o.name, o.cachepath, errs)
}

func (o *gcsFSObject) File() *os.File {
	return o.cachedcopy
}
func (o *gcsFSObject) Read(p []byte) (n int, err error) {
	return o.cachedcopy.Read(p)
}
func (o *gcsFSObject) Write(p []byte) (n int, err error) {
	return o.cachedcopy.Write(p)
}

func (o *gcsFSObject) Sync() error {

	if !o.opened {
		return fmt.Errorf("object isn't opened object:%s", o.name)
	}
	if o.readonly {
		return fmt.Errorf("trying to Sync a readonly object:%s", o.name)
	}

	var errs = make([]string, 0)

	cachedcopy, err := os.OpenFile(o.cachepath, os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("couldn't open localfile for sync'ing. local=%s err=%v",
			o.cachepath, err)
	}
	defer cachedcopy.Close()

	for try := 0; try < GCSRetries; try++ {
		if _, err := cachedcopy.Seek(0, os.SEEK_SET); err != nil {
			return fmt.Errorf("error seeking to start of cachedcopy err=%v", err) //don't retry on local filesystem errors
		}
		rd := bufio.NewReader(cachedcopy)

		wc := o.gcsb.Object(o.name).NewWriter(context.Background())

		if o.metadata != nil {
			wc.Metadata = o.metadata
			//contenttype is only used for viewing the file in a browser. (i.e. the GCS Object browser).
			ctype := ensureContextType(o.name, o.metadata)
			wc.ContentType = ctype
		}

		if _, err = io.Copy(wc, rd); err != nil {
			errs = append(errs, fmt.Sprintf("copy to remote object error:%v", err))
			err2 := wc.CloseWithError(err)
			if err2 != nil {
				errs = append(errs, fmt.Sprintf("CloseWithError error:%v", err2))
			}
			backoff(try)
			continue
		}

		if err = wc.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close gcs writer error:%v", err))
			backoff(try)
			continue
		}

		return nil
	}

	errmsg := strings.Join(errs, ",")
	return fmt.Errorf("GCS sync error after retry: (oname=%s cpath:%v) errors[%v]", o.name, o.cachepath, errmsg)
}

func (o *gcsFSObject) Close() error {
	if !o.opened {
		return nil
	}
	defer func() {
		os.Remove(o.cachepath)
		o.cachedcopy = nil
		o.opened = false
	}()

	serr := o.cachedcopy.Sync()
	cerr := o.cachedcopy.Close()
	if serr != nil || cerr != nil {
		return fmt.Errorf("error on sync and closing localfile. %s sync=%v, err=%v", o.cachepath, serr, cerr)
	}

	if o.opened && !o.readonly {
		err := o.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *gcsFSObject) Release() error {
	if o.cachedcopy != nil {
		o.cachedcopy.Close()
	}
	return os.Remove(o.cachepath)
}

//backoff sleeps a random amount so we can.
//retry failed requests using a randomized exponential backoff:
//wait a random period between [0..1] seconds and retry; if that fails,
//wait a random period between [0..2] seconds and retry; if that fails,
//wait a random period between [0..4] seconds and retry, and so on,
//with an upper bounds to the wait period being 16 seconds.
//http://play.golang.org/p/l9aUHgiR8J
func backoff(try int) {
	nf := math.Pow(2, float64(try))
	nf = math.Max(1, nf)
	nf = math.Min(nf, 16)
	r := rand.Int31n(int32(nf))
	d := time.Duration(r) * time.Second
	time.Sleep(d)
}
