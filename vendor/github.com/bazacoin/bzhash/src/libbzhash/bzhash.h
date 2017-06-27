/*
  This file is part of bzhash.

  bzhash is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  bzhash is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with bzhash.  If not, see <http://www.gnu.org/licenses/>.
*/

/** @file bzhash.h
* @date 2015
*/
#pragma once

#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stddef.h>
#include "compiler.h"

#define BZHASH_REVISION 23
#define BZHASH_DATASET_BYTES_INIT 1073741824U // 2**30
#define BZHASH_DATASET_BYTES_GROWTH 8388608U  // 2**23
#define BZHASH_CACHE_BYTES_INIT 1073741824U // 2**24
#define BZHASH_CACHE_BYTES_GROWTH 131072U  // 2**17
#define BZHASH_EPOCH_LENGTH 30000U
#define BZHASH_MIX_BYTES 128
#define BZHASH_HASH_BYTES 64
#define BZHASH_DATASET_PARENTS 256
#define BZHASH_CACHE_ROUNDS 3
#define BZHASH_ACCESSES 64
#define BZHASH_DAG_MAGIC_NUM_SIZE 8
#define BZHASH_DAG_MAGIC_NUM 0xFEE1DEADBADDCAFE

#ifdef __cplusplus
extern "C" {
#endif

/// Type of a seedhash/blockhash e.t.c.
typedef struct bzhash_h256 { uint8_t b[32]; } bzhash_h256_t;

// convenience macro to statically initialize an h256_t
// usage:
// bzhash_h256_t a = bzhash_h256_static_init(1, 2, 3, ... )
// have to provide all 32 values. If you don't provide all the rest
// will simply be unitialized (not guranteed to be 0)
#define bzhash_h256_static_init(...)			\
	{ {__VA_ARGS__} }

struct bzhash_light;
typedef struct bzhash_light* bzhash_light_t;
struct bzhash_full;
typedef struct bzhash_full* bzhash_full_t;
typedef int(*bzhash_callback_t)(unsigned);

typedef struct bzhash_return_value {
	bzhash_h256_t result;
	bzhash_h256_t mix_hash;
	bool success;
} bzhash_return_value_t;

/**
 * Allocate and initialize a new bzhash_light handler
 *
 * @param block_number   The block number for which to create the handler
 * @return               Newly allocated bzhash_light handler or NULL in case of
 *                       ERRNOMEM or invalid parameters used for @ref bzhash_compute_cache_nodes()
 */
bzhash_light_t bzhash_light_new(uint64_t block_number);
/**
 * Frees a previously allocated bzhash_light handler
 * @param light        The light handler to free
 */
void bzhash_light_delete(bzhash_light_t light);
/**
 * Calculate the light client data
 *
 * @param light          The light client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               an object of bzhash_return_value_t holding the return values
 */
bzhash_return_value_t bzhash_light_compute(
	bzhash_light_t light,
	bzhash_h256_t const header_hash,
	uint64_t nonce
);

/**
 * Allocate and initialize a new bzhash_full handler
 *
 * @param light         The light handler containing the cache.
 * @param callback      A callback function with signature of @ref bzhash_callback_t
 *                      It accepts an unsigned with which a progress of DAG calculation
 *                      can be displayed. If all goes well the callback should return 0.
 *                      If a non-zero value is returned then DAG generation will stop.
 *                      Be advised. A progress value of 100 means that DAG creation is
 *                      almost complete and that this function will soon return succesfully.
 *                      It does not mean that the function has already had a succesfull return.
 * @return              Newly allocated bzhash_full handler or NULL in case of
 *                      ERRNOMEM or invalid parameters used for @ref bzhash_compute_full_data()
 */
bzhash_full_t bzhash_full_new(bzhash_light_t light, bzhash_callback_t callback);

/**
 * Frees a previously allocated bzhash_full handler
 * @param full    The light handler to free
 */
void bzhash_full_delete(bzhash_full_t full);
/**
 * Calculate the full client data
 *
 * @param full           The full client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               An object of bzhash_return_value to hold the return value
 */
bzhash_return_value_t bzhash_full_compute(
	bzhash_full_t full,
	bzhash_h256_t const header_hash,
	uint64_t nonce
);
/**
 * Get a pointer to the full DAG data
 */
void const* bzhash_full_dag(bzhash_full_t full);
/**
 * Get the size of the DAG data
 */
uint64_t bzhash_full_dag_size(bzhash_full_t full);

/**
 * Calculate the seedhash for a given block number
 */
bzhash_h256_t bzhash_get_seedhash(uint64_t block_number);

#ifdef __cplusplus
}
#endif
