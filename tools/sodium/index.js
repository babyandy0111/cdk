// 主要用途：將 CDK 部署完所產生的環境變數內容寫入到 github 每個 repo 中
// usage: node tools/sodium/index.js "[environment_name]"

const sodium = require('tweetsodium');
const { Octokit } = require("@octokit/core");
require('dotenv').config();
const fs = require("fs");

// 目前只有 production
let envName = process.argv[2];
// 預設跟著 CDK 的 envName，但為了讓 CICD 自動抓符合的 branch 跑，env 名稱對應會不一樣
// CDK envName == production 時，github envName = master，這樣 CICD 跑到 master branch 時會自動抓名稱為 master 的環境設定
let githubEnvName = envName;
if (envName == "production") {
    githubEnvName = "master";
}

function encryptedValue (key, value) {
    // Convert the message and key to Uint8Array's (Buffer implements that interface)
    const messageBytes = Buffer.from(value);
    const keyBytes = Buffer.from(key, 'base64');


    // Encrypt using LibSodium.
    const encryptedBytes = sodium.seal(messageBytes, keyBytes);

    // Base64 the encrypted secret
    const encrypted = Buffer.from(encryptedBytes).toString('base64');

    return encrypted
}




// 可以操作的 repository
let usedRepos = ["api-automation-backend", "grpc-server", "apigateway"];
// 初始化 github
const github = new Octokit({auth: process.env.GITHUB_TOKEN})

// 讀取檔案內容並開始操作
let filename = "./" + envName + ".json";
fs.readFile(filename, "utf8", (err, data) => {;
    if (err) {
        console.error(err);
        return;
    }
    let input = JSON.parse(data);

    // 流程簡要說明：拉出所有 repo -> 判斷是否要加環境變數 -> 要加的話先行建立 environment -> 接著讀入 CDK 跑出來的 JSON 內容開始加入環境變數內容
    // 拉出所有 repository
    github.request("GET /orgs/"+process.env.GITHUB_ORGANIZATION+"/repos").then( obj => {
        for (let i in obj.data) {
            let d = obj.data[i];
            if (usedRepos.indexOf(d.name) < 0) {
                continue;
            }
            console.log("getting repo info:: >>  id: " + d.id + "; name: " + d.name);
            let createEnvUrl = "/repos/" + process.env.GITHUB_ORGANIZATION + "/" + d.name + "/environments/" + githubEnvName;
            console.log(createEnvUrl);
            let getPublicKeyUrl = "/repositories/"+d.id+"/environments/" + githubEnvName + "/secrets/public-key";
            // 建立環境
            github.request("PUT " + createEnvUrl, {
                deployment_branch_policy: {
                    protected_branches: false,
                    custom_branch_policies: true,
                }
            }).then(objEnv => {
                console.log("environment created / updated: " + objEnv.data.name);
                // 取得該 repository 下指定環境的 publickey
                github.request("GET " + getPublicKeyUrl).then(obj => {
                    console.log("get repo environment public_key:: >> id: " + d.id + "; name: " + d.name + "; public_key: " + obj.data.key);
                    // 開始寫入環境變數
                    for (let j in input) {
                        github.request("PUT /repositories/"+d.id+"/environments/" + githubEnvName +"/secrets/" + j, {
                            repository_id: d.id,
                            environment_name: githubEnvName,
                            secret_name: j,
                            encrypted_value: encryptedValue(obj.data.key, input[j]),
                            key_id: obj.data.key_id,
                        }).catch(err => {
                            console.log(err);
                        })
                    }
                }).catch(err => {
                    console.log("AAAA");
                    throw err;
                })
            }).catch(err => {
                console.log("BBBB");
                throw err;
            });
        }
    }).catch( err => {
        throw err;
    });
});