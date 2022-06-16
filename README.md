# inostar-heroku

### 部署方法
0. 去 inoreader 和 dropbox 里分别创建应用，获取环境变量里需要的 Token 或密钥
1. 先创建一个heroku账号

2. 安装heroku命令行 https://devcenter.heroku.com/articles/git
```shell
brew install heroku
```
3. 部署
```shell
git clone https://github.com/gonejack/inostar-heroku.git
cd inostar-heroku
heroku create -a example-app # heroku create -a 你的应用id
git push heroku main
```
4. 去设置环境变量
5. 访问 http://inostar.herokuapp.com/oauth2/auth 授权应用能访问你的 inoreader 数据，将页面返回的 TOKEN 的 JSON 填到应用的环境变量里
6. 每周需要重复一次步骤5，并更新环境变量里的记录，不更新会读取不了 inoreader 的 API.

### 环境变量
heroku 应用管理里需设置的环境变量
```dotenv
HOST=https://你的应用id.herokuapp.com

# 前往 https://www.dropbox.com/developers/apps 注册获得，需要勾选 files.metadata.write|read 和 files.content.write|read 权限
DROPBOX_TOKEN=你的dropbox访问Token

# 前往https://www.inoreader.com/starred#preferences-developer注册一个app获得以下信息
INOREADER_CLIENT_ID=你的clientID
INOREADER_CLIENT_SECRET=你的client密钥

# 访问auth返回的token
# 示例
# TOKEN={"access_token":"xxxxxxxxxxx","token_type":"Bearer","refresh_token":"xxxxxxxxxxx","expiry":"2021-12-13T13:15:23.678665265+08:00"}
TOKEN=访问auth返回的token

# 生成的.eml上的发件人信息
EML_FROM=gonejack@outlook.com
EML_TO=gonejack@outlook.com
# 生成.eml还是.eml.gz，填1则输出.eml.gz文件到dropbox中
EML_ZIP=0

LOG_LEVEL=info
TZ=Asia/Shanghai
```
