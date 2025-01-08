import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import type { FormProps } from "antd";
import { Button, Form, Input, Typography, Space, message, Switch } from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../utils/fetch";

const { Title } = Typography;

export default function NewPipeline() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [searchParams] = useSearchParams();
  const pipelineId = searchParams.get("id");
  const [testGitLoading, setTestGitLoading] = useState(false);
  const [errMsg, setErrMsg] = useState<string>("");
  const [testGitSuccess, setTestGitSuccess] = useState(false);

  useEffect(() => {
    if (pipelineId) {
      loadPipelineDetail();
    }
  }, [pipelineId]);

  type FieldType = {
    name: string;
    tag_template: string;
    envs: { key: string; val: string }[];
    use_git: boolean;
    repository?: string;
    branch?: string;
    username?: string;
    password?: string;
  };

  const onFinish: FormProps<FieldType>["onFinish"] = async (values) => {
    if (values.use_git && !testGitSuccess) {
      message.error("请先测试Git连接，并确保连接成功");
      return;
    }

    if (pipelineId) {
      const res = await fetchRequest("/api/update_pipeline/" + pipelineId, {
        method: "PUT",
        body: JSON.stringify(values),
      });
      message.success("更新成功");
      navigate("/new_pipeline/pipeline?id=" + res.id);
      return;
    }
    const res = await fetchRequest("/api/create_pipeline", {
      method: "POST",
      body: JSON.stringify(values),
    });
    message.success("创建成功");
    navigate("/new_pipeline/pipeline?id=" + res.id);
    console.log("Success:", values);
  };

  const onFinishFailed: FormProps<FieldType>["onFinishFailed"] = (
    errorInfo
  ) => {
    console.log("Failed:", errorInfo);
  };

  const loadPipelineDetail = async () => {
    const res = await fetchRequest("/api/pipeline/" + pipelineId, {
      method: "GET",
    });
    form.setFieldsValue(res);
  };

  const testGit = async () => {
    setTestGitLoading(true);
    setTestGitSuccess(false);
    setErrMsg("");
    const values = form.getFieldsValue();
    try {
      await fetchRequest("/api/test_git", {
        method: "POST",
        body: JSON.stringify(values),
      });
      setTestGitSuccess(true);
    } catch (e: any) {
      setErrMsg(e.message);
    } finally {
      setTestGitLoading(false);
    }
  };

  return (
    <div className="w-full bg-white ml-2 p-5">
      <Title level={4}>{pipelineId ? "编辑" : "新建"}Pipeline</Title>
      <Form
        name="basic"
        labelCol={{ span: 3 }}
        wrapperCol={{ span: 16 }}
        initialValues={{
          tag_template: "${TIMESTAMP}",
        }}
        style={{ maxWidth: 600, marginTop: "20px" }}
        onFinish={onFinish}
        onFinishFailed={onFinishFailed}
        autoComplete="off"
        form={form}
      >
        <Form.Item<FieldType>
          label="名称"
          name="name"
          rules={[{ required: true, message: "请输入名称" }]}
        >
          <Input placeholder="请输入名称" />
        </Form.Item>

        <Form.Item<FieldType>
          label="Tag模板"
          name="tag_template"
          extra="支持${COUNT}、${TIMESTAMP}"
          rules={[{ required: true, message: "请输入Tag模板" }]}
        >
          <Input placeholder="请输入Tag模板" />
        </Form.Item>

        <Form.Item label="环境变量">
          <Form.List name="envs">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key}>
                    <Form.Item
                      {...restField}
                      name={[name, "key"]}
                      rules={[{ required: true, message: "请输入key" }]}
                    >
                      <Input placeholder="Key" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, "val"]}
                      rules={[{ required: true, message: "请输入value" }]}
                    >
                      <Input placeholder="Value" />
                    </Form.Item>
                    <MinusCircleOutlined
                      onClick={() => {
                        remove(name);
                      }}
                    />
                  </Space>
                ))}
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    新增环境变量
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form.Item>

        <Form.Item label="使用Git" name="use_git" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.use_git !== currentValues.use_git
          }
        >
          {({ getFieldValue }) =>
            getFieldValue("use_git") && (
              <>
                <Form.Item
                  label="仓库地址"
                  name="repository"
                  rules={[{ required: true, message: "请输入仓库地址" }]}
                >
                  <Input placeholder="请输入仓库地址" />
                </Form.Item>

                <Form.Item
                  label="分支"
                  name="branch"
                  rules={[{ required: true, message: "请输入分支" }]}
                >
                  <Input placeholder="请输入分支" />
                </Form.Item>

                <Form.Item
                  label="用户名"
                  name="username"
                >
                  <Input placeholder="请输入用户名" />
                </Form.Item>

                <Form.Item
                  label="密码"
                  name="password"
                >
                  <Input placeholder="请输入密码" />
                </Form.Item>
                <Form.Item label={null}>
                  <Button onClick={testGit} loading={testGitLoading}>
                    测试连接
                  </Button>
                  {errMsg && <div className="text-red-500">{errMsg}</div>}
                  {testGitSuccess && <div className="text-green-500">连接成功</div>}
                </Form.Item>
              </>
            )
          }
        </Form.Item>

        <Form.Item label={null}>
          <Space>
            <Button type="primary" htmlType="submit">
              保存
            </Button>
            {pipelineId && (
              <Button
                color="default"
                variant="solid"
                onClick={() =>
                  navigate("/new_pipeline/step?id=" + searchParams.get("id"))
                }
              >
                创建Step
              </Button>
            )}
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}
