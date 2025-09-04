import React, { useState, useMemo, useCallback } from 'react';
// import { useNavigate } from 'react-router-dom';
import { Plus } from 'lucide-react';
import {
  useGetMoviesQuery,
  useGetQualityProfilesQuery,
  useGetRootFoldersQuery,
  useToggleMovieMonitorMutation,
  useSearchMovieMutation,
  useDeleteMovieMutation,
  useAddMovieMutation
} from '../store/api/radarrApi';
import { Button } from '../components/common';
import {
  MovieList,
  MovieDetail,
  SearchBar,
  AddMovieModal,
  type MovieFilter,
  type DiscoverMovie
} from '../components/movies';
import type { Movie, AddMovieRequest } from '../types/api';
import styles from './MoviesPage.module.css';

export const MoviesPage = () => {
  // Navigation hook available for future use
  // const navigate = useNavigate();
  const [view, setView] = useState<'grid' | 'list'>('grid');
  const [selectedMovies, setSelectedMovies] = useState<number[]>([]);
  const [selectedMovieId, setSelectedMovieId] = useState<number | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [sortBy, setSortBy] = useState('title');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [filter, setFilter] = useState<MovieFilter>({
    text: '',
    genres: [],
    years: { min: null, max: null },
    ratings: { min: null, max: null },
    status: [],
    monitored: null,
    hasFile: null,
    tags: []
  });

  // Queries
  const { data: movies = [], isLoading, error } = useGetMoviesQuery();
  const { data: qualityProfiles = [] } = useGetQualityProfilesQuery();
  const { data: rootFolders = [] } = useGetRootFoldersQuery();

  // Mutations
  const [toggleMonitor] = useToggleMovieMonitorMutation();
  const [searchMovie] = useSearchMovieMutation();
  const [deleteMovie] = useDeleteMovieMutation();
  const [addMovie] = useAddMovieMutation();

  // Filter and sort movies
  const filteredAndSortedMovies = useMemo(() => {
    const filtered = movies.filter((movie: Movie) => {
      // Text filter
      if (filter.text) {
        const searchText = filter.text.toLowerCase();
        if (!movie.title.toLowerCase().includes(searchText) &&
            !movie.originalTitle?.toLowerCase().includes(searchText) &&
            !movie.overview?.toLowerCase().includes(searchText)) {
          return false;
        }
      }

      // Genre filter
      if (filter.genres.length > 0) {
        if (!filter.genres.some(genre => movie.genres.includes(genre))) {
          return false;
        }
      }

      // Year filter
      if (filter.years.min !== null && movie.year < filter.years.min) {
        return false;
      }
      if (filter.years.max !== null && movie.year > filter.years.max) {
        return false;
      }

      // Rating filter
      if (filter.ratings.min !== null && movie.ratings.value < filter.ratings.min) {
        return false;
      }
      if (filter.ratings.max !== null && movie.ratings.value > filter.ratings.max) {
        return false;
      }

      // Status filter
      if (filter.status.length > 0) {
        if (!filter.status.includes(movie.status)) {
          return false;
        }
      }

      // Monitored filter
      if (filter.monitored !== null && movie.monitored !== filter.monitored) {
        return false;
      }

      // Has file filter
      if (filter.hasFile !== null && movie.hasFile !== filter.hasFile) {
        return false;
      }

      // Tags filter
      if (filter.tags.length > 0) {
        if (!filter.tags.some(tagId => movie.tags.includes(tagId))) {
          return false;
        }
      }

      return true;
    });

    // Sort movies
    filtered.sort((a: Movie, b: Movie) => {
      let aValue: string | number | Date;
      let bValue: string | number | Date;

      switch (sortBy) {
        case 'title':
          aValue = a.title.toLowerCase();
          bValue = b.title.toLowerCase();
          break;
        case 'year':
          aValue = a.year;
          bValue = b.year;
          break;
        case 'added':
          aValue = new Date(a.added).getTime();
          bValue = new Date(b.added).getTime();
          break;
        case 'ratings.value':
          aValue = a.ratings.value;
          bValue = b.ratings.value;
          break;
        case 'runtime':
          aValue = a.runtime;
          bValue = b.runtime;
          break;
        case 'status':
          aValue = a.status;
          bValue = b.status;
          break;
        default:
          aValue = a.title.toLowerCase();
          bValue = b.title.toLowerCase();
      }

      if (aValue < bValue) {
        return sortOrder === 'asc' ? -1 : 1;
      }
      if (aValue > bValue) {
        return sortOrder === 'asc' ? 1 : -1;
      }
      return 0;
    });

    return filtered;
  }, [movies, filter, sortBy, sortOrder]);

  // Get selected movie for detail view
  const selectedMovie = useMemo(() => {
    if (!selectedMovieId) return null;
    return movies.find(movie => movie.id === selectedMovieId) || null;
  }, [selectedMovieId, movies]);

  // Get unique genres for filtering
  const availableGenres = useMemo(() => {
    const genres = new Set<string>();
    movies.forEach(movie => {
      movie.genres.forEach(genre => genres.add(genre));
    });
    return Array.from(genres).sort();
  }, [movies]);

  // Handlers
  const handleSelectMovie = useCallback((movieId: number, selected: boolean) => {
    setSelectedMovies(prev =>
      selected
        ? [...prev, movieId]
        : prev.filter(id => id !== movieId)
    );
  }, []);

  const handleSelectAll = useCallback((selected: boolean) => {
    setSelectedMovies(selected ? filteredAndSortedMovies.map(movie => movie.id) : []);
  }, [filteredAndSortedMovies]);

  const handleToggleMonitor = useCallback(async (movieId: number, monitored: boolean) => {
    try {
      await toggleMonitor({ movieId, monitored }).unwrap();
    } catch (error) {
      console.error('Failed to toggle monitor:', error);
    }
  }, [toggleMonitor]);

  const handleSearchMovie = useCallback(async (movieId: number) => {
    try {
      await searchMovie(movieId).unwrap();
    } catch (error) {
      console.error('Failed to search movie:', error);
    }
  }, [searchMovie]);

  const handleDeleteMovie = useCallback(async (movieId: number) => {
    if (window.confirm('Are you sure you want to delete this movie?')) {
      try {
        await deleteMovie({ id: movieId, deleteFiles: false }).unwrap();
        setSelectedMovies(prev => prev.filter(id => id !== movieId));
      } catch (error) {
        console.error('Failed to delete movie:', error);
      }
    }
  }, [deleteMovie]);

  const handleMovieClick = useCallback((movieId: number) => {
    setSelectedMovieId(movieId);
  }, []);

  const handleCloseDetail = useCallback(() => {
    setSelectedMovieId(null);
  }, []);

  const handleSort = useCallback((field: string) => {
    if (sortBy === field) {
      setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(field);
      setSortOrder('asc');
    }
  }, [sortBy]);

  const handleBulkMonitor = useCallback(async (movieIds: number[], monitored: boolean) => {
    try {
      await Promise.all(
        movieIds.map(movieId =>
          toggleMonitor({ movieId, monitored }).unwrap()
        )
      );
      // Optionally clear selection after bulk operation
      setSelectedMovies([]);
    } catch (error) {
      console.error('Failed to bulk update monitor status:', error);
    }
  }, [toggleMonitor]);

  const handleBulkSearch = useCallback(async (movieIds: number[]) => {
    try {
      await Promise.all(
        movieIds.map(movieId =>
          searchMovie(movieId).unwrap()
        )
      );
      setSelectedMovies([]);
    } catch (error) {
      console.error('Failed to bulk search movies:', error);
    }
  }, [searchMovie]);

  const handleBulkDelete = useCallback(async (movieIds: number[]) => {
    if (window.confirm(`Are you sure you want to delete ${movieIds.length} movies?`)) {
      try {
        await Promise.all(
          movieIds.map(movieId =>
            deleteMovie({ id: movieId, deleteFiles: false }).unwrap()
          )
        );
        setSelectedMovies([]);
      } catch (error) {
        console.error('Failed to bulk delete movies:', error);
      }
    }
  }, [deleteMovie]);

  const handleAddMovie = useCallback(async (request: AddMovieRequest) => {
    try {
      await addMovie(request).unwrap();
    } catch (error) {
      console.error('Failed to add movie:', error);
      throw error;
    }
  }, [addMovie]);

  const handleSearchMovies = useCallback(async (): Promise<DiscoverMovie[]> => {
    try {
      // This would need to be implemented in the API
      // For now, return empty array
      return [];
    } catch (error) {
      console.error('Failed to search movies:', error);
      return [];
    }
  }, []);

  // Show movie detail if selected
  if (selectedMovie) {
    return (
      <MovieDetail
        movie={selectedMovie}
        onToggleMonitor={handleToggleMonitor}
        onSearch={handleSearchMovie}
        onEdit={(movieId) => {
          // Navigate to edit page when implemented
          console.log('Edit movie:', movieId);
        }}
        onDelete={handleDeleteMovie}
        onClose={handleCloseDetail}
      />
    );
  }

  if (error) {
    return (
      <div className={styles.container}>
        <div className={styles.error}>
          <h1>Error Loading Movies</h1>
          <p>Unable to load movies. Please try again later.</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.headerContent}>
          <div className={styles.title}>
            <h1>Movies</h1>
            <p>Manage your movie collection</p>
          </div>

          <Button
            onClick={() => setShowAddModal(true)}
            icon={<Plus size={16} />}
            className={styles.addButton}
          >
            Add Movies
          </Button>
        </div>

        <div className={styles.searchContainer}>
          <SearchBar
            filter={filter}
            onFilterChange={setFilter}
            availableGenres={availableGenres}
            availableStatuses={['Announced', 'In Cinemas', 'Released']}
          />
        </div>
      </div>

      <MovieList
        movies={filteredAndSortedMovies}
        loading={isLoading}
        error={error ? 'Failed to load movies' : undefined}
        view={view}
        onViewChange={setView}
        selectedMovies={selectedMovies}
        onSelectMovie={handleSelectMovie}
        onSelectAll={handleSelectAll}
        onToggleMonitor={handleToggleMonitor}
        onSearch={handleSearchMovie}
        onEdit={(movieId) => {
          // Navigate to edit page when implemented
          console.log('Edit movie:', movieId);
        }}
        onDelete={handleDeleteMovie}
        onMovieClick={handleMovieClick}
        sortBy={sortBy}
        sortOrder={sortOrder}
        onSort={handleSort}
        filterText={filter.text}
        onFilterChange={(text) => setFilter(prev => ({ ...prev, text }))}
        showBulkActions={selectedMovies.length > 0}
        onBulkMonitor={handleBulkMonitor}
        onBulkSearch={handleBulkSearch}
        onBulkDelete={handleBulkDelete}
      />

      <AddMovieModal
        isOpen={showAddModal}
        onClose={() => setShowAddModal(false)}
        onAddMovie={handleAddMovie}
        qualityProfiles={qualityProfiles}
        rootFolders={rootFolders}
        searchMovies={handleSearchMovies}
      />
    </div>
  );
};
