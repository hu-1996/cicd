import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { List } from "antd";
import Status from "./component/status";
import { fetchRequest } from "../../utils/fetch";

interface LoadDataParams {
  page: number;
  page_size: number;
}

export default function History() {
  const [searchParams] = useSearchParams();
  const pipelineId = searchParams.get("id");

  const [history, setHistory] = useState<any[]>([]);
  const [queryParams, setQueryParams] = useState<LoadDataParams>({
    page: 1,
    page_size: 10,
  });
  const [total, setTotal] = useState(0);

  useEffect(() => {
    if (pipelineId) {
      loadHostory();
    }
  }, [pipelineId, queryParams]);

  const loadHostory = async () => {
    const params = new URLSearchParams(
      Object.entries(queryParams).reduce((acc, [key, value]) => {
        acc[key] = String(value);
        return acc;
      }, {} as Record<string, string>)
    ).toString();
    const res = await fetchRequest(
      "/api/pipeline_jobs/" + pipelineId + "?" + params,
      {
        method: "GET",
      }
    );
    setTotal(res.total);
    setHistory(res.list);
  };

  return (
    <div>
      <List
        pagination={{
          pageSize: queryParams.page_size,
          current: queryParams.page,
          total: total,
          onChange: (page, pageSize) => {
            setQueryParams({
              ...queryParams,
              page,
              page_size: pageSize,
            });
          },
        }}
        dataSource={history}
        renderItem={(item) => (
          <List.Item extra={<Status steps={item.job_runners} />}>
            <List.Item.Meta
              title={item.tag}
              description={`触发时间：${item.updated_at}`}
            />
          </List.Item>
        )}
      />
    </div>
  );
}
