<template>
    <div id="app" onload="loopGetTaskList()">
        <el-container>
            <el-header>
                <div style="float: left;">
                    <el-button type="primary" @click="createTaskFormVisible = true">添加<i
                            class="el-icon-plus el-icon--right"></i></el-button>
                    <el-button type="success" @click="startAllTask">全部开始</el-button>
                    <el-button type="warning" @click="stopAllTask">全部暂停</el-button>
                </div>
            </el-header>
            <el-main>
                <el-table
                        header-cell-style="text-align:center;"
                        :data="taskData"
                        :default-sort = "{prop: 'Id', order: 'descending'}">
                    <el-table-column
                            prop="Id"
                            sortable
                            label="任务ID">
                    </el-table-column>
                    <el-table-column
                            prop="FileName"
                            label="文件名称">
                    </el-table-column>
                    <el-table-column
                            sortable
                            label="文件大小">
                        <template slot-scope="scope">
                            {{Math.ceil(scope.row.Size/1024/1024)}}MB
                        </template>
                    </el-table-column>
                    <el-table-column
                            prop="Speed"
                            sortable
                            label="下载速度">
                        <template slot-scope="scope">
                            <span v-if="scope.row.Speed/1024 > 1024">
                                {{Math.ceil(scope.row.Speed/1024/1024)}}Mb/s
                            </span>
                            <span v-else>
                                {{Math.ceil(scope.row.Speed/1024)}}kb/s
                            </span>
                        </template>
                    </el-table-column>
                    <el-table-column
                            sortable
                            label="下载进度">
                        <template slot-scope="scope">
                            {{scope.row.Progress}}%
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="已下载">
                        <template slot-scope="scope">
                            {{Math.ceil(scope.row.Downloaded/1024/1024)}}MB
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="操作">
                        <template slot-scope="scope">
                            <el-button type="success" icon="el-icon-download" circle @click="startTask(scope.row.Id)"></el-button>
                            <el-button type="warning" icon="el-icon-sort" circle @click="stopTask(scope.row.Id)"></el-button>
                            <el-button type="danger" icon="el-icon-delete" circle @click="deleteTask(scope.row.Id)"></el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-main>
        </el-container>
        <el-dialog title="添加任务" :visible.sync="createTaskFormVisible">
            <el-form :model="createTaskForm">
                <el-form-item label="下载链接">
                    <el-input v-model="createTaskForm.Url" auto-complete="off"></el-input>
                </el-form-item>
                <el-form-item label="保存路径">
                    <el-input v-model="createTaskForm.FilePath" auto-complete="off"></el-input>
                </el-form-item>
                <el-form-item label="并发数">
                    <el-select v-model="createTaskForm.PartCount"
                               placeholder="1">
                        <el-option label="2" value="2"></el-option>
                        <el-option label="4" value="4"></el-option>
                        <el-option label="8" value="8"></el-option>
                        <el-option label="16" value="16"></el-option>
                    </el-select>
                </el-form-item>
            </el-form>
            <div slot="footer" class="dialog-footer">
                <el-button @click="createTaskFormVisible = false">取 消</el-button>
                <el-button type="primary" @click="createTask">确 定</el-button>
            </div>
        </el-dialog>
    </div>
</template>
<script>
import axios from 'axios'
export default {
  name: 'app',
  data: function () {
    return {
      createTaskFormVisible: false,
      createTaskForm: {
        Url: '',
        FilePath: '',
        PartCount: undefined
      },
      taskData: []
    }
  },
  created: function () {

  },
  mounted: function () {
    let _this = this
    setInterval(() => {
      _this.getTaskList()
    }, 500)
  },
  methods: {
    getTaskList () {
      let _this = this
      axios.get('http://localhost:9988/task/get', {})
        .then(function (response) {
          if (response.status === 200) {
            console.log(response.data)
            _this.taskData = response.data
          }
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    createTask () {
      let _this = this
      let req = {
        PartCount: parseInt(this.createTaskForm.PartCount),
        FilePath: this.createTaskForm.FilePath,
        Url: this.createTaskForm.Url
      }
      parseInt(this.createTaskForm.PartCount)
      axios.post('http://localhost:9988/task/create', JSON.stringify(req))
        .then(function (response) {
          if (response.status === 200) {
            _this.createTaskFormVisible = false
            _this.$message({
              type: 'success',
              message: '新任务已加入下载队列!'
            })
          }
        })
        .catch(function (error) {
          return error
        })
    },
    startTask (id) {
      let _this = this
      axios.post('http://localhost:9988/task/start', JSON.stringify(id))
        .then(function (response) {
          if (response.status === 200) {
            _this.$message({
              type: 'success',
              message: '已开始下载!'
            })
          }
        })
        .catch(function () {
          _this.$message({
            type: 'danger',
            message: '已开始下载或下载出错!'
          })
        })
    },
    stopTask (id) {
      let _this = this
      axios.post('http://localhost:9988/task/stop', JSON.stringify(id))
        .then(function (response) {
          console.log(response.status === 200)
          if (response.status === 200) {
            _this.$message({
              type: 'warning',
              message: '已暂停下载!'
            })
          }
        })
        .catch(function () {
          _this.$message({
            type: 'danger',
            message: '已暂停!'
          })
        })
    },
    deleteTask (id) {
      let _this = this
      _this.$confirm('此操作将永久删除该文件, 是否继续?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        axios.post('http://localhost:9988/task/delete', JSON.stringify(id))
          .then(function (response) {
            if (response.status === 200) {
              _this.$message({
                type: 'success',
                message: '删除成功!'
              })
            }
          })
          .catch(function () {
            _this.$message({
              type: 'error',
              message: '删除失败'
            })
          })
      }).catch(() => {
        _this.$message({
          type: 'info',
          message: '已取消删除'
        })
      })
    },
    startAllTask () {
      let _this = this
      axios.post('http://localhost:9988/task/start_all')
        .then(function (response) {
          if (response.status === 200) {
            _this.$message({
              type: 'success',
              message: '已全部开始下载!'
            })
          }
        })
        .catch(function (error) {
          return error
        })
    },
    stopAllTask () {
      let _this = this
      axios.post('http://localhost:9988/task/stop_all')
        .then(function (response) {
          if (response.status === 200) {
            _this.$message({
              type: 'warning',
              message: '已全部暂停下载!'
            })
          }
        })
        .catch(function (error) {
          return error
        })
    }
  }
}
</script>

<style>
    #app {
        font-family: 'Avenir', Helvetica, Arial, sans-serif;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        text-align: center;
        color: #2c3e50;
        margin-top: 60px;
    }
</style>
