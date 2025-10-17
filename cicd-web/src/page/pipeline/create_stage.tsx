import { useEffect } from "react";
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
  Popconfirm,
} from "antd";
import { fetchRequest } from "../../utils/fetch";

const { Title } = Typography;

export default function NewStage() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [searchParams] = useSearchParams();

  const pipelineId = searchParams.get("id");
  const stageId = searchParams.get("stage_id");

  useEffect(() => {
    if (stageId) {
      loadStageDetail();
    }
  }, [stageId]);

  

  type FieldType = {
    pipeline_id: number;
    name: string;
    parallel: boolean;
  };

  const onFinish: FormProps<FieldType>["onFinish"] = async (values) => {
    if (!pipelineId) {
      message.error("缺少pipeline_id");
      return;
    }
    values.pipeline_id = Number(pipelineId);
    if (stageId) {
      await fetchRequest("/api/update_stage/" + stageId, {
        method: "PUT",
        body: JSON.stringify(values),
      });
      message.success("更新成功");
      return;
    }
    const res = await fetchRequest("/api/create_stage", {
      method: "POST",
      body: JSON.stringify(values),
    });
    message.success("创建成功");
    form.resetFields();
    navigate(
      "/new_pipeline/stage?id=" + pipelineId + "&stage_id=" + res.id
    );
  };

  const onFinishFailed: FormProps<FieldType>["onFinishFailed"] = (
    errorInfo
  ) => {
    console.log("Failed:", errorInfo);
  };

  const loadStageDetail = async () => {
    const res = await fetchRequest("/api/stage/" + stageId, {
      method: "GET",
    });
    form.setFieldsValue(res);
  };

  const confirm = async () => {
    await fetchRequest("/api/delete_stage/" + stageId, {
      method: "DELETE",
    });
    message.success("删除成功");
    navigate("/new_pipeline/pipeline?id=" + pipelineId);
  };

  return (
    <div className="w-full bg-white ml-2 p-5">
      <Title level={4}>{stageId ? "更新" : "新建"} Stage</Title>
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


        <Form.Item<FieldType> label="并行执行" name="parallel" required>
          <Switch checkedChildren="并行" unCheckedChildren="串行" />
        </Form.Item>

        <Form.Item label={null}>
          <Space>
            <Button type="primary" htmlType="submit">
              保存
            </Button>
            {stageId && (
              <>
                <Button
                  color="default"
                  variant="solid"
                  onClick={() => {
                    form.resetFields();
                    navigate("/new_pipeline/step?id=" + pipelineId + "&stage_id=" + stageId);
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
