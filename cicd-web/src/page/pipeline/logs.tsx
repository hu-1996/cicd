import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Descriptions,
  Select,
  Divider,
  Alert,
  Space,
  Button,
  message as msg,
} from "antd";
import type { DescriptionsProps } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../utils/fetch";

export default function Logs() {
  const [searchParams] = useSearchParams();
  const id = searchParams.get("id");

  const [job, setJob] = useState<any>();
  const [items, setItems] = useState<DescriptionsProps["items"]>([]);
  const [jobRunners, setJobRunners] = useState<any[]>([]);
  const [log, setLog] = useState<string>("");
  const [message, setMessage] = useState<string>("");

  useEffect(() => {
    loadDetail();
  }, [id]);

  useEffect(() => {
    if (job?.job_runner?.last_runner_id) {
      loadLogs(job?.job_runner?.last_runner_id);
    }
  }, [job?.job_runner?.last_runner_id]);

  const loadDetail = async () => {
    const res = await fetchRequest("/api/job_runner/" + id, {
      method: "GET",
    });
    setMessage(res?.job_runner?.message);
    const opts = res.job_runners.map((r: any) => ({
      value: r.last_runner_id,
      label: r.name + "@" + r.last_runner_id,
    }));
    setJobRunners(opts);
    setJob(res);
    setItems([
      {
        key: "1",
        label: "Tag",
        children: res?.job?.tag,
      },
      {
        key: "2",
        label: "执行机器",
        children: res?.job_runner?.assign_runners
          ?.map((r: any) => r.name)
          .join(", ") || '-',
      },
      {
        key: "3",
        label: "耗时",
        children: res?.job_runner?.cost,
      },
      {
        key: "4",
        label: "调度时间",
        children: res?.job_runner?.start_time,
      },
      {
        key: "5",
        label: "完成时间",
        children: res?.job_runner?.end_time,
      },
    ]);
  };

  const loadLogs = async (id: number) => {
    const res = await fetchRequest("/api/job_runner_log/" + id, {
      method: "GET",
    });
    setLog(res);
    if (job?.job_runner?.length > 0) {
      const rs = job?.job_runner?.filter((r: any) => r.id === id);
      if (rs.length > 0) {
        const message = rs[0].message;
        if (message) {
          setMessage(message);
        }
      } else {
        setMessage("");
      }
    }
  };

  const startStep = async (jobRunnerId: number) => {
    await fetchRequest("/api/start_job_step/" + jobRunnerId, {
      method: "POST",
    });
    msg.success("开始执行");
  };

  return (
    <div>
      <Descriptions title={job?.pipeline?.name} items={items} />
      <Divider />
      <Space>
        {job?.job_runner?.last_runner_id && (
          <Select
            options={jobRunners}
            className="w-[200px]"
            defaultValue={job?.job_runner?.last_runner_id}
            onChange={(val) => {
              loadLogs(val);
            }}
          />
        )}
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={() => startStep(job?.job_runner?.last_runner_id)}
        >
          重新执行
        </Button>
      </Space>
      {message && <Alert message={message} type="error" className="mt-4" />}
      <div
        className="overflow-y-scroll mt-4 bg-black text-white p-4 rounded-md"
        style={{ height: "calc(100vh - 300px)", minHeight: "300px" }}
      >
        <pre style={{ whiteSpace: "pre-wrap" }}>
          {(log || "").split("\n").map((line, i) => (
            <div key={i}>{line}</div>
          ))}
        </pre>
      </div>
    </div>
  );
}
