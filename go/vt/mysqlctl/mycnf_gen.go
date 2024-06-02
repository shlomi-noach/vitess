/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//  Generate my.cnf files from templates.

package mysqlctl

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"path"
	"strings"
	"text/template"

	"github.com/spf13/pflag"

	"vitess.io/vitess/go/vt/env"
	"vitess.io/vitess/go/vt/servenv"
)

// This files handles the creation of Mycnf objects for the default 'vt'
// file structure. These path are used by the mysqlctl commands.
//
// The default 'vt' file structure is as follows:
// - the root is specified by the environment variable VTDATAROOT,
//   and defaults to /vt
// - each tablet with uid NNNNNNNNNN is located in <root>/vt_NNNNNNNNNN
// - in that tablet directory, there is a my.cnf file for the mysql instance,
//   and 'data', 'innodb', 'relay-logs', 'bin-logs' directories.
// - these sub-directories might be symlinks to other places,
//   see comment for createTopDir, to allow some data types to be on
//   different disk partitions.

const (
	dataDir          = "data"
	innodbDir        = "innodb"
	relayLogDir      = "relay-logs"
	binLogDir        = "bin-logs"
	innodbDataSubdir = "innodb/data"
	innodbLogSubdir  = "innodb/logs"
)

var tabletDir string

func init() {
	for _, cmd := range []string{"mysqlctl", "mysqlctld", "vtcombo", "vttablet", "vttestserver", "vtctld", "vtctldclient"} {
		servenv.OnParseFor(cmd, registerMyCnfFlags)
	}
}

func registerMyCnfFlags(fs *pflag.FlagSet) {
	fs.StringVar(&tabletDir, "tablet_dir", tabletDir, "The directory within the vtdataroot to store vttablet/mysql files. Defaults to being generated by the tablet uid.")
}

// NewMycnf fills the Mycnf structure with vt root paths and derived values.
// This is used to fill out the cnfTemplate values and generate my.cnf.
// uid is a unique id for a particular tablet - it must be unique within the
// tabletservers deployed within a keyspace, lest there be collisions on disk.
// mysqldPort needs to be unique per instance per machine.
func NewMycnf(tabletUID uint32, mysqlPort int) *Mycnf {
	cnf := new(Mycnf)
	cnf.Path = MycnfFile(tabletUID)
	tabletDir := TabletDir(tabletUID)
	cnf.ServerID = tabletUID
	cnf.MysqlPort = mysqlPort
	cnf.DataDir = path.Join(tabletDir, dataDir)
	cnf.InnodbDataHomeDir = path.Join(tabletDir, innodbDataSubdir)
	cnf.InnodbLogGroupHomeDir = path.Join(tabletDir, innodbLogSubdir)
	cnf.SocketFile = path.Join(tabletDir, "mysql.sock")
	cnf.GeneralLogPath = path.Join(tabletDir, "general.log")
	cnf.ErrorLogPath = path.Join(tabletDir, "error.log")
	cnf.SlowLogPath = path.Join(tabletDir, "slow-query.log")
	cnf.RelayLogPath = path.Join(tabletDir, relayLogDir,
		fmt.Sprintf("vt-%010d-relay-bin", tabletUID))
	cnf.RelayLogIndexPath = cnf.RelayLogPath + ".index"
	cnf.RelayLogInfoPath = path.Join(tabletDir, relayLogDir, "relay-log.info")
	cnf.BinLogPath = path.Join(tabletDir, binLogDir,
		fmt.Sprintf("vt-%010d-bin", tabletUID))
	cnf.MasterInfoFile = path.Join(tabletDir, "master.info")
	cnf.PidFile = path.Join(tabletDir, "mysql.pid")
	cnf.TmpDir = path.Join(tabletDir, "tmp")
	// by default the secure-file-priv path is `tmp`
	cnf.SecureFilePriv = cnf.TmpDir
	return cnf
}

// TabletDir returns the default directory for a tablet
func TabletDir(uid uint32) string {
	if tabletDir != "" {
		return fmt.Sprintf("%s/%s", env.VtDataRoot(), tabletDir)
	}
	return DefaultTabletDirAtRoot(env.VtDataRoot(), uid)
}

// DefaultTabletDirAtRoot returns the default directory for a tablet given a UID and a VtDataRoot variable
func DefaultTabletDirAtRoot(dataRoot string, uid uint32) string {
	return fmt.Sprintf("%s/vt_%010d", dataRoot, uid)
}

// MycnfFile returns the default location of the my.cnf file.
func MycnfFile(uid uint32) string {
	return path.Join(TabletDir(uid), "my.cnf")
}

// TopLevelDirs returns the list of directories in the toplevel tablet directory
// that might be located in a different place.
func TopLevelDirs() []string {
	return []string{dataDir, innodbDir, relayLogDir, binLogDir}
}

// directoryList returns the list of directories to create in an empty
// mysql instance.
func (cnf *Mycnf) directoryList() []string {
	return []string{
		cnf.DataDir,
		cnf.InnodbDataHomeDir,
		cnf.InnodbLogGroupHomeDir,
		cnf.TmpDir,
		path.Dir(cnf.RelayLogPath),
		path.Dir(cnf.BinLogPath),
	}
}

// makeMycnf will substitute values
func (cnf *Mycnf) makeMycnf(partialcnf string) (string, error) {
	return cnf.fillMycnfTemplate(partialcnf)
}

// fillMycnfTemplate will fill in the passed in template with the values
// from Mycnf
func (cnf *Mycnf) fillMycnfTemplate(tmplSrc string) (string, error) {
	myTemplate, err := template.New("").Parse(tmplSrc)
	if err != nil {
		return "", err
	}
	var mycnfData strings.Builder
	err = myTemplate.Execute(&mycnfData, cnf)
	if err != nil {
		return "", err
	}
	return mycnfData.String(), nil
}

// RandomizeMysqlServerID generates a random MySQL server_id.
//
// The value assigned to ServerID will be in the range [100, 2^31):
// - It avoids 0 because that's reserved for mysqlbinlog dumps.
// - It also avoids 1-99 because low numbers are used for fake
// connections.  See NewBinlogConnection() in binlog/binlog_connection.go
// for more on that.
// - It avoids the 2^31 - 2^32-1 range, as there seems to be some
// confusion there. The main MySQL documentation at:
// http://dev.mysql.com/doc/refman/5.7/en/replication-options.html
// implies serverID is a full 32 bits number. The semi-sync log line
// at startup '[Note] Start semi-sync binlog_dump to slave ...'
// interprets the server_id as signed 32-bit (shows negative numbers
// for that range).
// Such an ID may also be responsible for a mysqld crash in semi-sync code,
// although we haven't been able to verify that yet. The issue for that is:
// https://github.com/vitessio/vitess/issues/2280
func (cnf *Mycnf) RandomizeMysqlServerID() error {
	// rand.Int(_, max) returns a value in the range [0, max).
	bigN, err := rand.Int(rand.Reader, big.NewInt(1<<31-100))
	if err != nil {
		return err
	}
	n := bigN.Uint64()
	// n is in the range [0, 2^31 - 100).
	// Add back 100 to put it in the range [100, 2^31).
	cnf.ServerID = uint32(n + 100)
	return nil
}
