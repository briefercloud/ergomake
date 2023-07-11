import Button from '../components/Button'
import List from '../components/List'
import { Repo } from '../hooks/useRepo'

type RepositoryListProps = {
  repos: Repo[]
  onConfigure: (repo: Repo) => void
}

type ConfigureButtonProps = {
  onClick: () => void
}

const ConfigureButton = ({ onClick }: ConfigureButtonProps) => {
  return <Button onClick={onClick}>Configure</Button>
}

const RepositoryList = ({ repos, onConfigure }: RepositoryListProps) => {
  const repoItems = repos.map((repo) => {
    const envWord = repo.environmentCount === 1 ? 'environment' : 'environments'

    return {
      name: repo.name,
      descriptionLeft: `${repo.environmentCount} ${envWord}`,
      descriptionRight: 'Deploys from GitHub',
      chevron: repo.lastDeployedAt ? undefined : (
        <ConfigureButton onClick={() => onConfigure(repo)} />
      ),
      url: repo.lastDeployedAt
        ? `/gh/${repo.owner}/repos/${repo.name}`
        : undefined,
      data: repo,
    }
  })

  return <List items={repoItems} />
}

export default RepositoryList
