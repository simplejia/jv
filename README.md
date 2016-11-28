# [jv](http://github.com/simplejia/jv)
## 实现初衷
* 用svn管理项目代码时，当基于svn分支做需求开发，新手常常出错，导致代码冲突不断，严重时还会影响主干代码的可靠性，因此我把最常用也是比较容易出错的svn命令封装起来，提供jv这个小工具来简化svn操作，同时也是一种代码流程上的规范

## 实践
* 以主干代码为主，只有主干代码才可以上到正式环境，主干svn路径规定如下：http://xxx/trunk/xxx，所有代码必须在trunk目录下
* 做需求时，应copy一个trunk分支(`jv -branch`)，分支代码放到你名下目录：http://xxx/branches/yourname/xxx，分支代码必须在branches/yourname目录下，创建分支同时，你得给一个简短的需求描述文字，jv工具会在列出所有分支时展现这段文字，以方便你记忆
* 当你的某个分支需要追上trunk最新代码时，你得通过`jv -catch`命令来实现，jv这个命令其实做了这几步：svn rm当前分支（别担心，不是真的删除）；svn copy一个同名分支；svn merge原分支修改；svn switch新分支。（千万不能简单把trunk代码直接merge到分支，这样很可能在把分支重新merge回trunk时造成不必要的冲突，另外svn提供的reintegrate参数方式其实也不能完全解决这个问题）
* 当你要把代码合到trunk时，你得切到trunk分支(`jv -switch`)，然后通过`jv -merge`命令来实现
* 当有多人开发同一需求时，可以建一个专门的分支用于合并所有人的最新代码用于测试

## 特性
* svn命令封装，不用输长长的一串，简单，安全，实用
* 可以在切换分支时，直接展示svn copy时带上的一段描述文字，方便你确定某一个分支究竟为什么需求做的

## 安装
> go get -u github.com/simplejia/jv

## 注意：
* 由于要识别url，请提前通过配置`JV_PATHS`环境变量来定义svn根路径，如：`http://xxx/proj1,http://xxx/proj2`，多个svn目录用`,`分隔，然后分支路径不包含`trunk`或是`branches`目录
* 默认用户名取的当前登录用户名，也可以通过配置`JV_USER`来指定用户名
* 目前对windows支持不好，对于svn命令的输出有可能是乱码(TODO)

## demo
```
$ jv
A partner for svn command
version: 1.7, Created by simplejia [11/2016]

Usage of jv:
  -branch
        新建branch
  -catch
        合并trunk最新修改
  -checkout
        获取主干代码
  -delbranch
        删除branch
  -merge
        选择一个branch来merge
  -switch
        列出所有branch以备switch

$ jv -switch

********************
本地有未提交的修改
********************
1.      /trunk  (master)
2.      qz      "圈子页，回帖页改造"
3.      sp      "视频列表输出"

choose branch: 
```
