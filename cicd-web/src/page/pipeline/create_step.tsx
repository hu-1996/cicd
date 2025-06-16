import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import type { FormProps } from "antd";
import {
  Button,
  Form,
  Input,
  Typography,
  Space,
  message,
  Switch,
  Row,
  Col,
  Popconfirm,
  Select,
} from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../utils/fetch";

const { Title } = Typography;

export default function NewPipeline() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [searchParams] = useSearchParams();

  const pipelineId = searchParams.get("id");
  const stepId = searchParams.get("step_id");

  const [runnerLabels, setRunnerLabels] = useState<string[]>([]);



  useEffect(() => {
    if (stepId) {
      loadStepDetail();
    }

    loadRunnerLabels()
  }, [stepId]);

  

  type FieldType = {
    pipeline_id: number;
    commands: string[];
    name: string;
    tag_template: string;
    trigger: string;
    trigger_policy: boolean;
    runner_label_match: string;
    multiple_runner_exec: boolean;
  };

  const onFinish: FormProps<FieldType>["onFinish"] = async (values) => {
    if (!pipelineId) {
      message.error("缺少pipeline_id");
      return;
    }
    if (values.commands?.length === 0) {
      message.error("缺少命令");
      return;
    }
    values.pipeline_id = Number(pipelineId);
    values.trigger = values.trigger_policy ? "auto" : "manual";
    if (searchParams.get("step_id")) {
      await fetchRequest("/api/update_step/" + stepId, {
        method: "PUT",
        body: JSON.stringify(values),
      });
      message.success("更新成功");
      location.reload();
      return;
    }
    const res = await fetchRequest("/api/create_step", {
      method: "POST",
      body: JSON.stringify(values),
    });
    message.success("创建成功");
    form.resetFields();
    navigate(
      "/new_pipeline/step?id=" + searchParams.get("id") + "&step_id=" + res.id
    );
  };

  const onFinishFailed: FormProps<FieldType>["onFinishFailed"] = (
    errorInfo
  ) => {
    console.log("Failed:", errorInfo);
  };

  const loadStepDetail = async () => {
    const res = await fetchRequest("/api/step/" + stepId, {
      method: "GET",
    });
    res.trigger_policy = res.trigger !== "manual";
    form.setFieldsValue(res);
  };

  const loadRunnerLabels = async () => {
    const res = await fetchRequest("/api/list_runner_label", {
      method: "GET",
    });
    setRunnerLabels(res);
  };

  const confirm = async () => {
    await fetchRequest("/api/delete_step/" + stepId, {
      method: "DELETE",
    });
    message.success("删除成功");
    navigate("/new_pipeline/step?id=" + searchParams.get("id"));
  };

  return (
    <div className="w-full bg-white ml-2 p-5">
      <Title level={4}>{pipelineId ? "更新" : "新建"} Step</Title>
      <Form
        name="basic"
        labelCol={{ span: 6 }}
        wrapperCol={{ span: 16 }}
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

        <Form.Item label="命令" required>
          <Form.List name="commands">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Row key={key} gutter={10}>
                    <Col span={22}>
                      <Form.Item
                        {...restField}
                        name={[name]}
                        rules={[{ required: true, message: "请输入命令" }]}
                      >
                        <Input.TextArea placeholder="请输入命令" />
                      </Form.Item>
                    </Col>
                    <Col span={2}>
                      <MinusCircleOutlined onClick={() => remove(name)} />
                    </Col>
                  </Row>
                ))}
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    新增命令
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form.Item>

        <Form.Item<FieldType> label="触发策略" name="trigger_policy" required>
          <Switch checkedChildren="自动" unCheckedChildren="手动" />
        </Form.Item>

        <Form.Item<FieldType>
          label="执行机器标签"
          name="runner_label_match"
          rules={[{ required: true, message: "请输入执行机器标签" }]}
        >
          <Select placeholder="请输入执行机器标签" options={runnerLabels.map((label) => ({ label, value: label }))} />
        </Form.Item>

        <Form.Item<FieldType>
          label="允许多机器执行"
          name="multiple_runner_exec"
          required
        >
          <Switch />
        </Form.Item>

        <Form.Item label={null}>
          <Space>
            <Button type="primary" htmlType="submit">
              保存
            </Button>
            {stepId && (
              <>
                <Button
                  color="default"
                  variant="solid"
                  onClick={() => {
                    form.resetFields();
                    navigate("/new_pipeline/step?id=" + searchParams.get("id"));
                  }}
                >
                  创建Step
                </Button>
                <Popconfirm
                  title="提示"
                  description={`是否删除?`}
                  onConfirm={() => confirm()}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button danger>删除</Button>
                </Popconfirm>
              </>
            )}
          </Space>
        </Form.Item>
      </Form>
    </div>
  );
}
