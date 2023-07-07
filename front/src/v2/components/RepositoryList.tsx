import { Repo } from '../../hooks/useRepo'
import List from '../components/List'

type RepositoryListProps = {
  repos: Repo[]
}

const RepositoryList = ({ repos }: RepositoryListProps) => {
  const repoItems = repos.map((repo) => {
    const envWord = repo.environmentCount === 1 ? 'environment' : 'environments'
    return {
      name: repo.name,
      descriptionLeft: `${repo.environmentCount} ${envWord}`,
      descriptionRight: 'Deploys from GitHub',
      url: `/v2/gh/${repo.owner}/repos/${repo.name}`,
    }
  })

  return <List items={repoItems} />
}

export default RepositoryList
